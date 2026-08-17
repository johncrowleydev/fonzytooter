export function Dashboard() {
  return (
    <div className="space-y-8">
      <section>
        <p className="mb-2 text-sm font-medium uppercase tracking-[0.2em] text-slate-500">
          Home
        </p>
        <h1 className="text-3xl font-semibold tracking-tight">Learn at your own pace.</h1>
        <p className="mt-3 max-w-2xl text-slate-400">
          The scaffold is running. Curriculum, review scheduling, exercises, progress, and real tutor
          providers will be added incrementally as the learning workflow takes shape.
        </p>
      </section>

      <section className="grid gap-4 md:grid-cols-3">
        <DashboardCard title="Reviews" value="—" detail="FSRS not wired yet" />
        <DashboardCard title="Continue" value="Orientation" detail="Curriculum shell ready" />
        <DashboardCard title="Objectives" value="—" detail="Mastery model comes next" />
      </section>

      <section className="rounded-2xl border border-slate-800 bg-slate-900/50 p-6">
        <h2 className="text-lg font-medium">Tutor</h2>
        <p className="mt-2 max-w-2xl text-sm leading-6 text-slate-400">
          The tutor button is mounted globally at the bottom-right. The streaming API shape is already
          connected; the current backend provider intentionally reports that inference is not configured.
        </p>
      </section>
    </div>
  )
}

type DashboardCardProps = {
  title: string
  value: string
  detail: string
}

function DashboardCard({ title, value, detail }: DashboardCardProps) {
  return (
    <div className="rounded-2xl border border-slate-800 bg-slate-900/50 p-5">
      <div className="text-sm text-slate-400">{title}</div>
      <div className="mt-2 text-2xl font-semibold">{value}</div>
      <div className="mt-1 text-xs text-slate-500">{detail}</div>
    </div>
  )
}
