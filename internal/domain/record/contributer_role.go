package record

import "fmt"

type ContributerRole string

const (
	ContributerRoleAuthor         ContributerRole = "author"
	ContributerRoleInterpreter    ContributerRole = "interpreter"
	ContributerRoleCompiler       ContributerRole = "compiler"
	ContributerRoleCompilerAuthor ContributerRole = "compiler_author"
)

const (
	ContributerRoleIdAuthor         = "RL000010"
	ContributerRoleIdInterpreter    = "RL000020"
	ContributerRoleIdCompiler       = "RL000030"
	ContributerRoleIdCompilerAuthor = "RL000031"
)

var roleToIdMap = map[ContributerRole]string{
	ContributerRoleAuthor:         ContributerRoleIdAuthor,
	ContributerRoleInterpreter:    ContributerRoleIdInterpreter,
	ContributerRoleCompiler:       ContributerRoleIdCompiler,
	ContributerRoleCompilerAuthor: ContributerRoleIdCompilerAuthor,
}

var idToRoleMap = map[string]ContributerRole{
	ContributerRoleIdAuthor:         ContributerRoleAuthor,
	ContributerRoleIdInterpreter:    ContributerRoleInterpreter,
	ContributerRoleIdCompiler:       ContributerRoleCompiler,
	ContributerRoleIdCompilerAuthor: ContributerRoleCompilerAuthor,
}

// GetId 貢献者役割のIDを返却する
func (r ContributerRole) GetId() (string, error) {
	id, ok := roleToIdMap[r]
	if !ok {
		return "", fmt.Errorf("invalid role id: %s", r)
	}

	return id, nil
}

// RoleFromId 貢献者役割IDをもとに貢献者役割を返却する
func RoleFromId(id string) (ContributerRole, error) {
	role, ok := idToRoleMap[id]
	if !ok {
		return "", fmt.Errorf("invalid role id: %s", id)
	}

	return role, nil
}

// IsValid 役割が正しい値であればtrueを返却する
func (r ContributerRole) IsValid() bool {
	switch r {
	case ContributerRoleAuthor,
		ContributerRoleInterpreter,
		ContributerRoleCompiler,
		ContributerRoleCompilerAuthor:
		return true
	default:
		return false
	}
}
