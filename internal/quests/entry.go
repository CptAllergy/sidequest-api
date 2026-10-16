package quests

import "time"

// This file contains the entry types and their associated interfaces for input and response handling.

// Entry represents a single entry in a quest, which can be of different types (text, image, link, etc.)
type Entry struct {
	Id        string
	QuestId   string
	UserId    string
	Type      EntryType
	Content   EntryContent
	Icon      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type EntryUserInput struct {
	Type    EntryType
	Content EntryContentUserInput
}

type EntryContent interface {
	entryType() EntryType
}

// EntryContentUserInput Inputs can be used for creating or updating user entries through the API
type EntryContentUserInput interface {
	entryInput()
}

// EntryContentResponse Responses are used for returning data to the client, and may have different fields than the input types
type EntryContentResponse interface {
	entryContentResponse()
}

type EntryTextContent struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
}

func (EntryTextContent) entryType() EntryType {
	return Text
}

func (EntryTextContent) entryContentResponse() {}
func (EntryTextContent) entryInput()           {}

type EntryImageContent struct {
	Title       string `json:"title"`
	ImageKey    string `json:"imageKey" validate:"required"`
	Description string `json:"description"`
}

func (EntryImageContent) entryType() EntryType {
	return Image
}

func (EntryImageContent) entryInput() {}

type EntryImageContentResponse struct {
	Title       string `json:"title"`
	ImageUrl    string `json:"imageUrl"`
	Description string `json:"description"`
}

func (EntryImageContentResponse) entryContentResponse() {}

type EntryLinkContent struct {
	Title       string `json:"title"`
	Url         string `json:"url" validate:"required"`
	Description string `json:"description"`
}

func (EntryLinkContent) entryType() EntryType {
	return Link
}
func (EntryLinkContent) entryContentResponse() {}
func (EntryLinkContent) entryInput()           {}

type EntryStatusChangeContent struct {
	OldStatus string `json:"oldStatus"`
	NewStatus string `json:"newStatus"`
}

func (EntryStatusChangeContent) entryType() EntryType {
	return StatusChange
}

func (EntryStatusChangeContent) entryContentResponse() {}
