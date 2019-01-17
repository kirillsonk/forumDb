package models

import "time"

// easyjson:json
type Post struct {
	Author   string    `json:"author"`
	Created  time.Time `json:"created"`
	Forum    string    `json:"forum"`
	Id       int64     `json:"id"`
	IsEdited bool      `json:"isEdited"`
	Message  string    `json:"message"`
	Parent   int64     `json:"parent"`
	Thread   int32     `json:"thread"`
}

// easyjson:json
type PostList []Post

// easyjson:json
type Thread struct {
	Author  string    `json:"author"`
	Created time.Time `json:"created"`
	Forum   string    `json:"forum"`
	Id      int32     `json:"id"`
	Message string    `json:"message"`
	Slug    string    `json:"slug"`
	Title   string    `json:"title"`
	Votes   int32     `json:"votes"`
}

// easyjson:json
type ThreadList []Thread

// easyjson:json
type Forum struct {
	Posts   int64  `json:"posts"`
	Slug    string `json:"slug"`
	Threads int32  `json:"threads"`
	Title   string `json:"title"`
	User    string `json:"user"`
}

// easyjson:json
type ForumList []Forum

// easyjson:json
type User struct {
	About    string `json:"about"`
	Email    string `json:"email"`
	FullName string `json:"fullname"`
	NickName string `json:"nickname"`
}

// easyjson:json
type UserList []User

// easyjson:json
type PostDetail struct {
	Author *User   `json:"author"`
	Forum  *Forum  `json:"forum"`
	Post   *Post   `json:"post"`
	Thread *Thread `json:"thread"`
}

// easyjson:json
type Vote struct {
	Nickname string `json:"nickname"`
	Voice    int32  `json:"voice"`
	Thread   string `json:"-"`
}

// easyjson:json
type VoteList []Vote

// easyjson:json
type Status struct {
	Forum  int32 `json:"forum"`
	Post   int64 `json:"post"`
	Thread int32 `json:"thread"`
	User   int32 `json:"user"`
}

// easyjson:json
type Error struct {
	Message string `json:"message"`
}
