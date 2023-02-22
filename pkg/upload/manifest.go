package upload

import (
	"fmt"
	core2 "github.com/pennsieve/pennsieve-go-core/pkg/core"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dbTable"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest/manifestFile"
	log "github.com/sirupsen/logrus"
	"sync"
)

var syncWG sync.WaitGroup

const batchSize = 25 // maximum batch size for batchPut action on dynamodb
const nrWorkers = 2  // preliminary profiling shows that more workers don't improve efficiency for up to 1000 files

type ManifestSession struct {
	FileTableName string
	TableName     string
	Client        core2.DynamoDBAPI
	SNSClient     core2.SnsAPI
	SNSTopic      string
	S3Client      core2.S3API
}

// fileWalk channel used to distribute FileDTOs to the workers importing the files in DynamoDB
type fileWalk chan manifestFile.FileDTO

// AddFiles manages the workers and defines the go routines to add files to upload db.
func (s ManifestSession) AddFiles(manifestId string, items []manifestFile.FileDTO, forceStatus *manifestFile.Status) *manifest.AddFilesStats {

	// Populate DynamoDB with concurrent workers.
	walker := make(fileWalk, batchSize)
	result := make(chan manifest.AddFilesStats, nrWorkers)

	// List crawler
	go func() {
		// Gather the files to upload by walking the path recursively
		defer func() {
			close(walker)
		}()
		log.WithFields(
			log.Fields{
				"manifest_id": manifestId,
			},
		).Debug(fmt.Sprintf("Adding %d number of items from upload.", len(items)))

		for _, f := range items {
			walker <- f
		}
	}()

	// Initiate a set of upload sync workers as go-routines
	for w := 1; w <= nrWorkers; w++ {
		w2 := int32(w)
		syncWG.Add(1)
		log.Debug("starting worker:", w2)

		go func() {
			stats, _ := s.createOrUpdateFile(w2, walker, manifestId, forceStatus)
			result <- *stats
			defer func() {
				log.Debug("Closing Worker: ", w2)
				syncWG.Done()
			}()
		}()
	}

	syncWG.Wait()
	close(result)

	resp := manifest.AddFilesStats{}
	for r := range result {
		resp.NrFilesUpdated += r.NrFilesUpdated
		resp.NrFilesRemoved += r.NrFilesRemoved
		resp.FileStatus = append(resp.FileStatus, r.FileStatus...)
		resp.FailedFiles = append(resp.FailedFiles, r.FailedFiles...)
	}

	return &resp

}

// createOrUpdateFile is run in a goroutine and grabs set of files from channel and calls updateDynamoDb.
func (s ManifestSession) createOrUpdateFile(workerId int32, files fileWalk, manifestId string, forceStatus *manifestFile.Status) (*manifest.AddFilesStats, error) {

	response := manifest.AddFilesStats{}

	tbl := dbTable.ManifestFileTable{}

	// Create file slice of size "batchSize" or smaller if end of list.
	var fileSlice []manifestFile.FileDTO = nil
	for record := range files {
		fileSlice = append(fileSlice, record)

		// When the number of items in fileSize matches the batchSize --> make call to update dynamodb
		if len(fileSlice) == batchSize {
			stats, _ := tbl.SyncFiles(s.Client, s.FileTableName, manifestId, fileSlice, forceStatus)
			fileSlice = nil

			response.NrFilesUpdated += stats.NrFilesUpdated
			response.NrFilesRemoved += stats.NrFilesRemoved
			response.FailedFiles = append(response.FailedFiles, stats.FailedFiles...)
			response.FileStatus = append(response.FileStatus, stats.FileStatus...)
		}
	}

	// Add final partially filled fileSlice to database
	if fileSlice != nil {
		stats, _ := tbl.SyncFiles(s.Client, s.FileTableName, manifestId, fileSlice, forceStatus)
		response.NrFilesUpdated += stats.NrFilesUpdated
		response.NrFilesRemoved += stats.NrFilesRemoved
		response.FailedFiles = append(response.FailedFiles, stats.FailedFiles...)
		response.FileStatus = append(response.FileStatus, stats.FileStatus...)
	}

	return &response, nil
}
