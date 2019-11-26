package view

type Creator interface {
	CreateTable() error
}

type Dropper interface {
	DropTable() error
}

type View interface {
	Creator
	Dropper
}
