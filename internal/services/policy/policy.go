package policy

// "/": ["admin", "student", "teacher"]
type AccessControl struct {
	List map[string][]string
}

func New() *AccessControl {
	return &AccessControl{make(map[string][]string)}
}

func (ac *AccessControl) Add(url string, roles ...string) {
	ac.List[url] = roles
}

func (ac *AccessControl) Contains(k, v string) bool {
	for _, r := range ac.List[k] {
		if r == v {
			return true
		}
	}
	return false
}
