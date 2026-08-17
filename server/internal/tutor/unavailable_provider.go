package tutor

import "context"

type UnavailableProvider struct{}

func NewUnavailableProvider() *UnavailableProvider {
	return &UnavailableProvider{}
}

func (p *UnavailableProvider) StreamTurn(ctx context.Context, _ TurnRequest) (<-chan Event, error) {
	events := make(chan Event, 2)

	go func() {
		defer close(events)

		select {
		case <-ctx.Done():
			return
		case events <- Event{
			Type: EventTextDelta,
			Text: "The tutor UI and streaming contract are wired, but no inference provider is configured yet.",
		}:
		}

		select {
		case <-ctx.Done():
			return
		case events <- Event{Type: EventCompleted}:
		}
	}()

	return events, nil
}
