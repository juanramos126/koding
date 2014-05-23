package models

import (
	"bytes"
	"fmt"
	"html/template"
	"net/url"
	"socialapi/config"
	// socialmodels "socialapi/models"
	// "socialapi/workers/notification/models"
	"time"
)

type TemplateParser struct {
	UserContact *UserContact
}

var (
	mainTemplateFile        = "../socialapi/workers/emailnotifier/templates/main.tmpl"
	footerTemplateFile      = "../socialapi/workers/emailnotifier/templates/footer.tmpl"
	contentTemplateFile     = "../socialapi/workers/emailnotifier/templates/content.tmpl"
	gravatarTemplateFile    = "../socialapi/workers/emailnotifier/templates/gravatar.tmpl"
	groupTemplateFile       = "../socialapi/workers/emailnotifier/templates/group.tmpl"
	previewTemplateFile     = "../socialapi/workers/emailnotifier/templates/preview.tmpl"
	objectTemplateFile      = "../socialapi/workers/emailnotifier/templates/object.tmpl"
	unsubscribeTemplateFile = "../socialapi/workers/emailnotifier/templates/unsubscribe.tmpl"
)

func NewTemplateParser() *TemplateParser {
	return &TemplateParser{}
}

func (tp *TemplateParser) RenderInstantTemplate(mc *MailerContainer) (string, error) {
	if err := tp.validateTemplateParser(); err != nil {
		return "", err
	}

	ec, err := buildEventContent(mc)
	if err != nil {
		return "", err
	}
	content := tp.renderContentTemplate(ec, mc)

	return tp.renderTemplate(mc.Content.TypeConstant, content, mc.CreatedAt)
}

func (tp *TemplateParser) RenderDailyTemplate(containers []*MailerContainer) (string, error) {
	if err := tp.validateTemplateParser(); err != nil {
		return "", err
	}

	var contents string
	for _, mc := range containers {
		ec, err := buildEventContent(mc)
		if err != nil {
			continue
		}
		c := tp.renderContentTemplate(ec, mc)
		contents = c + contents
	}

	return tp.renderTemplate("daily", contents, time.Now())
}

func (tp *TemplateParser) validateTemplateParser() error {
	if tp.UserContact == nil {
		return fmt.Errorf("TemplateParser UserContact is not set")
	}

	return nil
}

func (tp *TemplateParser) renderTemplate(contentType, content string, date time.Time) (string, error) {
	t := template.Must(template.ParseFiles(
		mainTemplateFile, footerTemplateFile, unsubscribeTemplateFile))
	mc := tp.buildMailContent(contentType, getMonthAndDay(date))

	mc.Content = template.HTML(content)

	var doc bytes.Buffer
	if err := t.Execute(&doc, mc); err != nil {
		return "", err
	}

	return doc.String(), nil
}

func (tp *TemplateParser) buildMailContent(contentType string, currentDate string) *MailContent {
	return &MailContent{
		FirstName:   tp.UserContact.FirstName,
		CurrentDate: currentDate,
		Unsubscribe: &UnsubscribeContent{
			Token:       tp.UserContact.Token,
			ContentType: contentType,
			Recipient:   url.QueryEscape(tp.UserContact.Email),
		},
		Uri: config.Get().Uri,
	}
}

func buildEventContent(mc *MailerContainer) (*EventContent, error) {
	ec := &EventContent{
		Action:       mc.ActivityMessage,
		ActivityTime: prepareTime(mc),
		Uri:          config.Get().Uri,
		Slug:         mc.Slug,
		Message:      mc.Message,
		Group:        mc.Group,
		ObjectType:   mc.ObjectType,
		Size:         20,
	}

	actor, err := FetchUserContact(mc.Activity.ActorId)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while retrieving actor details", err)
	}
	ec.ActorContact = *actor

	return ec, nil
}

func appendGroupTemplate(t *template.Template, mc *MailerContainer) {
	var groupTemplate *template.Template
	if mc.Group.Name == "" || mc.Group.Slug == "koding" {
		groupTemplate = getEmptyTemplate()
	} else {
		groupTemplate = template.Must(
			template.ParseFiles(groupTemplateFile))
	}

	t.AddParseTree("group", groupTemplate.Tree)
}

func (tp *TemplateParser) renderContentTemplate(ec *EventContent, mc *MailerContainer) string {
	t := template.Must(template.ParseFiles(contentTemplateFile, gravatarTemplateFile))
	appendPreviewTemplate(t, mc)
	appendGroupTemplate(t, mc)

	buf := bytes.NewBuffer([]byte{})
	t.ExecuteTemplate(buf, "content", ec)

	return buf.String()
}

func appendPreviewTemplate(t *template.Template, mc *MailerContainer) {
	var previewTemplate, objectTemplate *template.Template
	if mc.Message == "" {
		previewTemplate = getEmptyTemplate()
		objectTemplate = getEmptyTemplate()
	} else {
		previewTemplate = template.Must(template.ParseFiles(previewTemplateFile))
		objectTemplate = template.Must(template.ParseFiles(objectTemplateFile))
	}

	t.AddParseTree("preview", previewTemplate.Tree)
	t.AddParseTree("object", objectTemplate.Tree)
}

func getEmptyTemplate() *template.Template {
	return template.Must(template.New("").Parse(""))
}

func getMonthAndDay(t time.Time) string {
	return t.Format("Jan 02")
}

func prepareDate(mc *MailerContainer) string {
	return mc.Activity.CreatedAt.Format("Jan 02")
}

func prepareTime(mc *MailerContainer) string {
	return mc.Activity.CreatedAt.Format("15:04")
}
