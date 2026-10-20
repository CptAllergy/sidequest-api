package quests

import (
	"encoding/json"
	"fmt"

	db "github.com/cptallergy/sidequest-api/internal/db/sqlc"
)

func toQuestResponse(q Quest) QuestResponse {
	return QuestResponse{
		Id:          q.Id,
		UserId:      q.UserId,
		Title:       q.Title,
		Description: q.Description,
		Type:        q.Type,
		Status:      q.Status,
		ImageUrl:    q.ImageUrl,
		CreatedAt:   q.CreatedAt,
		UpdatedAt:   q.UpdatedAt,
	}
}

func toQuestResponses(qs []Quest) []QuestResponse {
	results := make([]QuestResponse, 0, len(qs))
	for _, q := range qs {
		r := toQuestResponse(q)
		results = append(results, r)
	}
	return results
}

func toEntryResponse(e Entry) (EntryResponse, error) {
	// TODO need to map image key to url on imageType
	content, ok := e.Content.(EntryContentResponse)
	if !ok {
		return EntryResponse{}, fmt.Errorf("error decoding entry content")
	}

	return EntryResponse{
		Id:        e.Id,
		QuestId:   e.QuestId,
		UserId:    e.UserId,
		Type:      e.Type,
		Content:   content,
		Icon:      e.Icon,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}, nil
}

func toEntryResponses(es []Entry) ([]EntryResponse, error) {
	results := make([]EntryResponse, 0, len(es))
	for _, e := range es {
		r, err := toEntryResponse(e)
		if err != nil {
			return []EntryResponse{}, err
		}
		results = append(results, r)
	}
	return results, nil
}

func (e *CreateEntryRequest) decodeEntryInput() (EntryUserInput, error) {
	var (
		content EntryContentUserInput
		err     error
	)
	switch e.Type {
	case Text:
		content, err = decodeContentInput[EntryTextContent](e.Content)
	case Image:
		content, err = decodeContentInput[EntryImageContent](e.Content)
	case Link:
		content, err = decodeContentInput[EntryLinkContent](e.Content)
	default:
		return EntryUserInput{}, fmt.Errorf("unknown or unhandled entry type: %s", e.Type)
	}
	if err != nil {
		return EntryUserInput{}, fmt.Errorf("error decoding entry content: %v", err)
	}

	return EntryUserInput{
		Type:    e.Type,
		Content: content,
	}, nil
}

func toQuestFromRow(q db.Quest, imageUrl *string) Quest {
	return Quest{
		Id:          q.ID.String(),
		UserId:      q.UserID,
		Title:       q.Title,
		Description: q.Description,
		// TODO validate these 2
		Type:      QuestType(q.Type),
		Status:    StatusType(q.Status),
		ImageUrl:  imageUrl,
		CreatedAt: q.CreatedAt.Time,
		UpdatedAt: q.UpdatedAt.Time,
	}
}

func toEntryFromRow(e db.QuestEntry) (Entry, error) {
	var (
		content EntryContent
		err     error
	)
	switch EntryType(e.Type) {
	case Text:
		content, err = decodeContent[EntryTextContent](e.Content)
	case Image:
		content, err = decodeContent[EntryImageContent](e.Content)
	case Link:
		content, err = decodeContent[EntryLinkContent](e.Content)
	case StatusChange:
		content, err = decodeContent[EntryStatusChangeContent](e.Content)
	default:
		return Entry{}, fmt.Errorf("unknown or unhandled entry type: %s", e.Type)
	}
	if err != nil {
		return Entry{}, fmt.Errorf("error decoding entry content: %v", err)
	}

	return Entry{
		Id:      e.ID.String(),
		QuestId: e.QuestID.String(),
		UserId:  e.UserID,
		Type:    EntryType(e.Type),
		Content: content,
		// TODO add icons to db
		Icon:      "default",
		CreatedAt: e.CreatedAt.Time,
		UpdatedAt: e.UpdatedAt.Time,
	}, nil
}

func decodeContentInput[T EntryContentUserInput](raw json.RawMessage) (EntryContentUserInput, error) {
	var c T
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, err
	}
	return c, nil
}

func decodeContent[T EntryContent](b []byte) (EntryContent, error) {
	var c T
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return c, nil
}

func toCursorFromQuest(q Quest) QuestCursor {
	return QuestCursor{
		UpdatedAt: q.UpdatedAt,
		Id:        q.Id,
	}
}

func toCursorFromEntry(e Entry) EntryCursor {
	return EntryCursor{
		CreatedAt: e.CreatedAt,
		Id:        e.Id,
	}
}
