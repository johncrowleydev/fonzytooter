package tutor

import "context"

type UnavailableProvider struct{}

func NewUnavailableProvider() *UnavailableProvider {
	return &UnavailableProvider{}
}

func (p *UnavailableProvider) Stream(ctx context.Context, _ ModelRequest) (<-chan ProviderEvent, error) {
	events := make(chan ProviderEvent, 2)

	go func() {
		defer close(events)

		select {
		case <-ctx.Done():
			return
		case events <- ProviderEvent{
			Type: ProviderEventTextDelta,
			Text: "The tutor UI and streaming contract are wired, but no inference provider is configured yet.",
		}:
		}

		select {
		case <-ctx.Done():
			return
		case events <- ProviderEvent{Type: ProviderEventCompleted}:
		}
	}()

	return events, nil
}
