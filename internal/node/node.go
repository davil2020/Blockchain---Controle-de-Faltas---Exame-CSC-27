package node

const (
	RoleProfessor = "PROFESSOR"
	RoleAluno     = "ALUNO"
	RoleDAE       = "DAE"
)

type Node struct {
	ID    string
	Role  string
	Peers []string
}

func NewNode(id, role string) *Node {
	return &Node{
		ID:   id,
		Role: role,
	}
}

func (n *Node) CanMineBlocks() bool {
	return n.Role == RoleProfessor || n.Role == RoleDAE
}

func (n *Node) CanRegisterAttendance() bool {
	return n.Role == RoleProfessor
}

func (n *Node) CanAddJustifications() bool {
	return n.Role == RoleDAE
}

func (n *Node) CanViewFullChain() bool {
	return n.Role == RoleProfessor || n.Role == RoleDAE
}
