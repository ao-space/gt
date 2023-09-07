package request

// Menu represents a menu item.
type Menu struct {
	Path      string    `json:"path"`
	Name      string    `json:"name"`
	Component string    `json:"component"`
	Redirect  string    `json:"redirect,omitempty"`
	Meta      MetaProps `json:"meta"`
	Children  []Menu    `json:"children,omitempty"`
}

// MetaProps represents the meta properties of a menu item.
type MetaProps struct {
	Icon        string `json:"icon"`
	Title       string `json:"title"`
	IsLink      string `json:"isLink"`
	IsHide      bool   `json:"isHide"`
	IsFull      bool   `json:"isFull"`
	IsAffix     bool   `json:"isAffix"`
	IsKeepAlive bool   `json:"isKeepAlive"`
}
