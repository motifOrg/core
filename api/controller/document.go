package controller

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	model "core/internal/model"
	service "core/internal/service"
	logger "core/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetDocuments godoc
//
//	@Summary		List documents
//	@Description	Lists thesis documents uploaded to a campaign
//	@Tags			documents
//	@Produce		json
//	@Param			id	path		string	true	"Campaign ID"
//	@Success		200	{array}		model.Document
//	@Router			/campaigns/{id}/documents [get]
func (c *Controller) GetDocuments(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	documents, err := c.documents.List(id)
	if err != nil {
		logger.Error(
			"Failed to list documents: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}

	context.JSON(
		http.StatusOK,
		gin.H{"data": documents},
	)
}

// UploadDocument godoc
//
//	@Summary		Upload document
//	@Description	Stores a PDF or presentation and its metadata. Content is not parsed yet.
//	@Tags			documents
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			id		path		string	true	"Campaign ID"
//	@Param			file	formData	file	true	"PDF, PPT, PPTX or KEY"
//	@Success		201		{object}	model.Document
//	@Failure		400		{object}	ErrorResponse
//	@Failure		413		{object}	ErrorResponse
//	@Router			/campaigns/{id}/documents [post]
func (c *Controller) UploadDocument(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	header, err := context.FormFile("file")
	if err != nil {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": "file is required (form field: file)"},
		)

		return
	}

	file, err := header.Open()
	if err != nil {
		logger.Error(
			"Failed to open upload: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "failed to read upload"},
		)

		return
	}
	defer file.Close()

	head := make([]byte, 512)
	read, _ := io.ReadFull(file, head)
	head = head[:read]

	mimeType, err := c.storage.Validate(header.Filename, header.Size, head)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, service.ErrFileTooLarge) {
			status = http.StatusRequestEntityTooLarge
		}
		context.JSON(
			status,
			gin.H{"error": fmt.Sprintf("%s: %s", header.Filename, err.Error())},
		)
		return
	}

	key, err := c.storage.Save(
		id,
		header.Filename,
		io.MultiReader(bytes.NewReader(head), file),
	)
	if err != nil {
		logger.Error(
			"Failed to store upload: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "failed to store file"},
		)

		return
	}

	document := model.Document{
		CampaignID: id,
		Filename:   filepath.Base(header.Filename),
		StorageKey: key,
		MimeType:   mimeType,
		SizeBytes:  header.Size,
		Status:     model.DocumentStatusUploaded,
	}
	if err := c.documents.Create(&document);
		err != nil {
		_ = c.storage.Delete(key)

		logger.Error(
			"Failed to save document: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}

	context.JSON(
		http.StatusCreated,
		gin.H{"data": document},
	)
}

// DeleteDocument godoc
//
//	@Summary		Delete document
//	@Description	Removes the document record and the stored file
//	@Tags			documents
//	@Produce		json
//	@Param			id	path		string	true	"Document ID"
//	@Success		200	{object}	map[string]interface{}
//	@Failure		404	{object}	ErrorResponse
//	@Router			/documents/{id} [delete]
func (c *Controller) DeleteDocument(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	document, err := c.documents.Get(id)
	if err != nil {
		logger.Error(
			"Failed to load document: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}

	if err := c.documents.Delete(id);
		err != nil {
		logger.Error(
			"Failed to delete document: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}

	if err := c.storage.Delete(document.StorageKey);
		err != nil {
		logger.Error(
			"Failed to remove stored file %s: %v",
			document.StorageKey,
			err,
		)
	}

	context.JSON(
		http.StatusOK,
		gin.H{"data": gin.H{"id": id}},
	)
}
