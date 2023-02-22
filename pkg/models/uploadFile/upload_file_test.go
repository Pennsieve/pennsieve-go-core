package uploadFile

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHandler(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
	){
		"sorts upload files by path":      testSortFiles,
		"correctly maps files to folders": testFolderMapping,
		"test ignore leading slash":       testRemoveLeadingTrailingSlash,
	} {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
		})
	}
}

func testSortFiles(t *testing.T) {

	uploadFile1 := UploadFile{
		ManifestId: "",
		Path:       "folder1/asd/123/",
		Name:       "",
		Extension:  "",
		Type:       0,
		SubType:    "",
		Icon:       0,
		Size:       0,
		ETag:       "",
	}
	uploadFile2 := UploadFile{
		ManifestId: "",
		Path:       "folder1/asd/123",
		Name:       "",
		Extension:  "",
		Type:       0,
		SubType:    "",
		Icon:       0,
		Size:       0,
		ETag:       "",
	}

	uploadFiles := []UploadFile{
		uploadFile1,
		uploadFile2,
	}

	var u UploadFile
	u.Sort(uploadFiles)
	assert.Equal(t, uploadFiles[0], uploadFile2)

}

func testFolderMapping(t *testing.T) {

	uploadFile1 := UploadFile{
		ManifestId: "",
		Path:       "folder1/folder2/folder3",
		Name:       "",
		Extension:  "",
		Type:       0,
		SubType:    "",
		Icon:       0,
		Size:       0,
		ETag:       "",
	}
	uploadFile2 := UploadFile{
		ManifestId: "",
		Path:       "folder1/folder10",
		Name:       "",
		Extension:  "",
		Type:       0,
		SubType:    "",
		Icon:       0,
		Size:       0,
		ETag:       "",
	}
	uploadFile3 := UploadFile{
		ManifestId: "",
		Path:       "folder2/folder1/folder8",
		Name:       "",
		Extension:  "",
		Type:       0,
		SubType:    "",
		Icon:       0,
		Size:       0,
		ETag:       "",
	}

	uploadFiles := []UploadFile{
		uploadFile1,
		uploadFile2,
		uploadFile3,
	}

	var u UploadFile
	folderMap := u.GetUploadFolderMap(uploadFiles, "")

	// Number of folders
	assert.Equal(t, 7, len(folderMap))

	// Check Folder exists
	assert.True(t, folderMap["folder1/folder10"] != nil)
	assert.True(t, folderMap["folder1/unknownFolder"] == nil)

	// Check Folder Parents
	assert.Equal(t, folderMap["folder1/folder10"].ParentNodeId, folderMap["folder1"].NodeId)
	assert.Equal(t, folderMap["folder1/folder2/folder3"].ParentNodeId, folderMap["folder1/folder2"].NodeId)
	assert.Equal(t, folderMap["folder1/folder10"].ParentNodeId, folderMap["folder1/folder2"].ParentNodeId)

	// Check folder depth
	assert.Equal(t, 0, folderMap["folder1"].Depth)
	assert.Equal(t, 2, folderMap["folder1/folder2/folder3"].Depth)

	// Check population of children in parents
	assert.Contains(t, folderMap["folder1"].Children, folderMap["folder1/folder10"])
	assert.Contains(t, folderMap["folder1"].Children, folderMap["folder1/folder2"])
	assert.NotContains(t, folderMap["folder1"].Children, folderMap["folder2/folder1"])

	//** Check with alternative root folder.

	folderMap2 := u.GetUploadFolderMap(uploadFiles, "hello/you")

	// Number of folders
	assert.Equal(t, 9, len(folderMap2))

	// Check Folder exists
	assert.True(t, folderMap2["hello/you/folder1/folder10"] != nil)
	assert.True(t, folderMap2["hello/you/folder1/unknownFolder"] == nil)

	// Check Folder Parents
	assert.Equal(t, folderMap2["hello/you/folder1/folder10"].ParentNodeId, folderMap2["hello/you/folder1"].NodeId)
	assert.Equal(t, folderMap2["hello/you/folder1/folder2/folder3"].ParentNodeId, folderMap2["hello/you/folder1/folder2"].NodeId)
	assert.Equal(t, folderMap2["hello/you/folder1/folder10"].ParentNodeId, folderMap2["hello/you/folder1/folder2"].ParentNodeId)

	// Check folder depth
	assert.Equal(t, 2, folderMap2["hello/you/folder1"].Depth)
	assert.Equal(t, 4, folderMap2["hello/you/folder1/folder2/folder3"].Depth)

}

func testRemoveLeadingTrailingSlash(t *testing.T) {
	uploadFile1 := UploadFile{
		ManifestId: "",
		Path:       "/folder1/folder2",
		Name:       "",
		Extension:  "",
		Type:       0,
		SubType:    "",
		Icon:       0,
		Size:       0,
		ETag:       "",
	}
	uploadFile2 := UploadFile{
		ManifestId: "",
		Path:       "/////folder1/folder10",
		Name:       "",
		Extension:  "",
		Type:       0,
		SubType:    "",
		Icon:       0,
		Size:       0,
		ETag:       "",
	}
	uploadFile3 := UploadFile{
		ManifestId: "",
		Path:       "/folder1/folder10///",
		Name:       "",
		Extension:  "",
		Type:       0,
		SubType:    "",
		Icon:       0,
		Size:       0,
		ETag:       "",
	}
	uploadFile4 := UploadFile{
		ManifestId: "",
		Path:       "/folder1/folder10/",
		Name:       "",
		Extension:  "",
		Type:       0,
		SubType:    "",
		Icon:       0,
		Size:       0,
		ETag:       "",
	}

	uploadFiles := []UploadFile{
		uploadFile1,
		uploadFile2,
		uploadFile3,
		uploadFile4,
	}

	var u UploadFile
	folderMap := u.GetUploadFolderMap(uploadFiles, "")

	t.Log(folderMap)
	// Number of folders
	assert.Equal(t, 3, len(folderMap))

}
