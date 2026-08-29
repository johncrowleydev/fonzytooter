#!/usr/bin/env bash

set -Eeuo pipefail

if [[ $# -ne 5 ]]; then
  echo "usage: jc-dev.sh COMMIT_SHA RUN_ID RUN_ATTEMPT ARCHIVE SHA256" >&2
  exit 2
fi

commit_sha=$1
run_id=$2
run_attempt=$3
archive=$4
archive_sha256=$5

[[ $commit_sha =~ ^[0-9a-f]{40}$ ]]
[[ $run_id =~ ^[0-9]+$ ]]
[[ $run_attempt =~ ^[0-9]+$ ]]
[[ $archive =~ ^/tmp/helix-academy-[0-9]+-[0-9]+\.tar\.gz$ ]]
[[ $archive_sha256 =~ ^[0-9a-f]{64}$ ]]

release_id="${commit_sha}-${run_id}-${run_attempt}"
app_root=/opt/helix-academy
web_root=/var/www/helix.johncrowley.dev/html
app_release="${app_root}/releases/${release_id}"
web_release="${web_root}/releases/${release_id}"
app_incoming="${app_root}/releases/.incoming-${release_id}"
web_incoming="${web_root}/releases/.incoming-${release_id}"
stage=$(mktemp -d /tmp/helix-academy-deploy.XXXXXX)
activated=false

cleanup() {
  case $stage in
    /tmp/helix-academy-deploy.*) rm -rf -- "$stage" ;;
  esac
  case $archive in
    /tmp/helix-academy-[0-9]*-[0-9]*.tar.gz) rm -f -- "$archive" ;;
  esac
}

restore_link() {
  local root=$1
  local previous=$2
  local temporary="${root}/current.rollback"

  sudo rm -f -- "$temporary"
  if [[ -n $previous ]]; then
    sudo ln -s "$previous" "$temporary"
    sudo mv -Tf "$temporary" "${root}/current"
  else
    sudo rm -f -- "${root}/current"
  fi
}

rollback() {
  local status=$?
  trap - ERR
  set +e
  if [[ $activated == true ]]; then
    echo "Deployment failed after activation; restoring the previous release." >&2
    restore_link "$app_root" "$previous_app"
    restore_link "$web_root" "$previous_web"
    sudo systemctl restart helix-academy.service
  fi
  sudo rm -rf -- "$app_incoming" "$web_incoming"
  exit "$status"
}

trap cleanup EXIT
trap rollback ERR

echo "${archive_sha256}  ${archive}" | sha256sum --check --strict
tar --extract --gzip --file "$archive" --directory "$stage"

package="${stage}/deploy-package"
test -x "${package}/helix-academy"
test -f "${package}/web/index.html"
test -f "${package}/curriculum/sources.yaml"
test ! -e "$app_release"
test ! -e "$web_release"
test ! -e "$app_incoming"
test ! -e "$web_incoming"

sudo install -d -o root -g root -m 0755 \
  "$app_incoming" \
  "${app_incoming}/curriculum" \
  "$web_incoming"
sudo install -o root -g root -m 0755 \
  "${package}/helix-academy" \
  "${app_incoming}/helix-academy"
sudo cp -a "${package}/curriculum/." "${app_incoming}/curriculum/"
sudo cp -a "${package}/web/." "$web_incoming/"
sudo chown -R root:root "$app_incoming" "$web_incoming"
sudo chmod -R u=rwX,go=rX "$app_incoming" "$web_incoming"
sudo mv "$app_incoming" "$app_release"
sudo mv "$web_incoming" "$web_release"

previous_app=$(readlink "${app_root}/current" || true)
previous_web=$(readlink "${web_root}/current" || true)
sudo rm -f -- "${app_root}/current.next" "${web_root}/current.next"
sudo ln -s "releases/${release_id}" "${app_root}/current.next"
sudo ln -s "releases/${release_id}" "${web_root}/current.next"
activated=true
sudo mv -Tf "${app_root}/current.next" "${app_root}/current"
sudo mv -Tf "${web_root}/current.next" "${web_root}/current"

sudo systemctl restart helix-academy.service
healthy=false
for _ in $(seq 1 30); do
  if curl --fail --silent --show-error http://127.0.0.1:18082/api/health >/dev/null; then
    healthy=true
    break
  fi
  sleep 1
done
if [[ $healthy != true ]]; then
  sudo systemctl --no-pager --full status helix-academy.service >&2 || true
  sudo journalctl -u helix-academy.service -n 100 --no-pager >&2 || true
  false
fi

activated=false
trap - ERR

if [[ -d /srv/jc-dev/apps/helix-academy/.git ]]; then
  git -C /srv/jc-dev/apps/helix-academy fetch --quiet origin "$commit_sha" &&
    git -C /srv/jc-dev/apps/helix-academy checkout --quiet --detach "$commit_sha" ||
    echo "Warning: deployed successfully, but the server source checkout was not updated." >&2
fi

echo "Deployed ${commit_sha} as ${release_id}."
