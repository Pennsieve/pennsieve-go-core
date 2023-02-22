package uploadFolder

// UploadFolder represents a folder that is part of an upload session.
type UploadFolder struct {
	Id           int64           // Id of the folder
	NodeId       string          // NodeId of the folder
	Name         string          // Name of the folder
	ParentId     int64           // Id of the parent (-1 for root)
	ParentNodeId string          // NodeId for the parent ("" for root)
	Depth        int             // Depth of folder in relation to root
	Children     []*UploadFolder // Children contains folders that need to be created that have current folder as parent.
}

// UploadFolderMap maps path to UploadFolder
type UploadFolderMap = map[string]*UploadFolder
