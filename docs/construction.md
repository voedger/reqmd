# Construction requirements

- The tool shall be implemented in Go
- All files but main.go shall be in `internal` folder and its subfolders
- Design of the solution shall follow SOLID principles
  - Tracing shall be abstracted by ITracer interface, implemented by Tracer
  - All necessary interfaces shall be injected into Tracer during construction (NewTracer)
- Naming
  - Interface names shall start with `I`
  - Interface implementation names shall be deduced from the interface name by removing the I prefix and possibly lowercasing the first letter
  - All interfaces shall be defined in a separate file `interfaces.go`
  - All data structures used across the application shall be defined in the `models.go` file
- "github.com/go-git/go-git/v5" shall be used for git operations