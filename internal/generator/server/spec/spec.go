package spec

// servers

type Server struct {
	Services []*Service
}

type Implement struct {
	Template string
	Path     string
}

type Aggregation struct {
	Repository *Repo
	Others     []*Service
}
type Repository struct {
	Repository *Repo
	Others     []*Service
}
type Write struct {
	Server  *Implement
	Command *Command
	Request *Request
	Sub     *Sub
	Job     *Job
	Others  []*Service
}
type Read struct {
	Server    *Implement
	Aggregate *Aggregate
	Inquiry   *Inquiry
	Others    []*Service
}

// service

type Service struct {
	Template string
	Path     string
	Methods  []*Method
}

type Repo struct {
	Template string
	Path     string
	Get      *Method
	List     *Method
	Create   *Method
	Delete   *Method
	Update   *Method
	Others   []*Method
}

type Command struct {
	Template string
	Path     string
	Methods  []*Method
}
type Request struct {
	Template string
	Path     string
	Methods  []*Method
}
type Sub struct {
	Template string
	Path     string
	Methods  []*Method
}
type Job struct {
	Template string
	Path     string
	Methods  []*Method
}

type Aggregate struct {
	Template string
	Path     string
	Methods  []*Method
}
type Inquiry struct {
	Template string
	Path     string
	List     *Method
	Detail   *Method
	Others   []*Method
}

// method

type Method struct {
	Name string
	In   string
	Out  string
}

// spec
type Spec struct {
	Dist        string
	Aggregation *Aggregation
	Repository  *Repository
	Write       *Write
	Read        *Read
	Servers     []*Server
}

func Load(path string) (*Spec, error) {
	return &Spec{
		Aggregation: &Aggregation{
			Repository: &Repo{
				Template: "aggregation",
				Path:     "aggregation/aggregation.go",
			},
		},
		Repository: &Repository{
			Repository: &Repo{
				Template: "repository",
				Path:     "repository/repository.go",
			},
		},
		Write: &Write{
			Server: &Implement{
				Template: "write.server",
				Path:     "write/server.go",
			},
			Request: &Request{
				Template: "write.request",
				Path:     "write/request.go",
			},
		},
	}, nil
}
