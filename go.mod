module golang.org/x/mobile

go 1.23.0

// The modern go/types type checker produces types.Alias
// types for the explicit representation of type aliases.
// (Initial opt-in support for it was added in Go 1.22,
// and it became the default behavior in Go 1.23.)
//
// TODO(go.dev/issue/70698): Update the golang.org/x/mobile/bind
// code generator for the new behavior and delete this temporary¹
// forced pre-1.23 go/types behavior.
//
// ¹ It's temporary because this godebug setting will be removed
//   in a future Go release.
godebug gotypesalias=0

require (
	golang.org/x/exp/shiny v0.0.0-20230817173708-d852ddb80c63
	golang.org/x/image v0.27.0
	golang.org/x/mod v0.24.0
	golang.org/x/sync v0.14.0
	golang.org/x/tools v0.33.0
)

require golang.org/x/sys v0.33.0 // indirect
