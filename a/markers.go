package a

import "github.com/anathema-framework/component"

type Service interface {
	component.Marker
	service()
}

type Provider interface {
	component.Marker
	provider()
}

