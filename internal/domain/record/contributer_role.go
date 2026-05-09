package record

type ContributerRole string

const (
	ContributerRoleAuthor         = "author"
	ContributerRoleInterpreter    = "interpreter"
	ContributerRoleCompiler       = "compiler"
	ContributerRoleCompilerAuthor = "compiler_author"
)

func (s ContributerRole) IsValid() bool {
	switch s {
	case ContributerRoleAuthor,
		ContributerRoleInterpreter,
		ContributerRoleCompiler,
		ContributerRoleCompilerAuthor:
		return true
	default:
		return false
	}
}
