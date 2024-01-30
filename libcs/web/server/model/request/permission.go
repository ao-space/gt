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
	Icon        string `json:"icon"`        //element-plus icon name
	Title       string `json:"title"`       //menu item title
	IsLink      string `json:"isLink"`      //if isLink is not empty, the menu item will be a link to the given url
	IsHide      bool   `json:"isHide"`      //if isHide is true, the menu item will be hidden
	IsFull      bool   `json:"isFull"`      //if isFull is true, the menu item will be displayed in full screen
	IsAffix     bool   `json:"isAffix"`     //if isAffix is true, the menu item cannot be closed on the tab line
	IsKeepAlive bool   `json:"isKeepAlive"` //if isKeepAlive is true, the menu item will be cached
}
