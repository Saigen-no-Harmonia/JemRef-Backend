package record

type ContributorRole string

const (
	ContributorRoleAuthor         ContributorRole = "author"
	ContributorRoleInterpreter    ContributorRole = "interpreter"
	ContributorRoleCompiler       ContributorRole = "compiler"
	ContributorRoleCompilerAuthor ContributorRole = "compiler_author"
)

const (
	ContributorRoleIdAuthor         = "RL000010"
	ContributorRoleIdInterpreter    = "RL000020"
	ContributorRoleIdCompiler       = "RL000030"
	ContributorRoleIdCompilerAuthor = "RL000031"
)

var roleToIdMap = map[ContributorRole]string{
	ContributorRoleAuthor:         ContributorRoleIdAuthor,
	ContributorRoleInterpreter:    ContributorRoleIdInterpreter,
	ContributorRoleCompiler:       ContributorRoleIdCompiler,
	ContributorRoleCompilerAuthor: ContributorRoleIdCompilerAuthor,
}

var idToRoleMap = map[string]ContributorRole{
	ContributorRoleIdAuthor:         ContributorRoleAuthor,
	ContributorRoleIdInterpreter:    ContributorRoleInterpreter,
	ContributorRoleIdCompiler:       ContributorRoleCompiler,
	ContributorRoleIdCompilerAuthor: ContributorRoleCompilerAuthor,
}

// GetId 貢献者役割のIDを返却する
func (r ContributorRole) GetId() string {
	id, ok := roleToIdMap[r]
	if !ok {
		panic("invalid role")
	}

	return id
}

// RoleFromId 貢献者役割IDをもとに貢献者役割を返却する
func RoleFromId(id string) ContributorRole {
	role, ok := idToRoleMap[id]
	if !ok {
		panic("invalid role id")
	}

	return role
}

// IsValid 役割が正しい値であればtrueを返却する
func (r ContributorRole) IsValid() bool {
	switch r {
	case ContributorRoleAuthor,
		ContributorRoleInterpreter,
		ContributorRoleCompiler,
		ContributorRoleCompilerAuthor:
		return true
	default:
		return false
	}
}
