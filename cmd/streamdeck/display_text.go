package main

import (
	"context"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func init() {
	registerDisplayElement("text", &displayElementText{})
}

type displayElementText struct {
	running bool
}

func (d displayElementText) Display(ctx context.Context, idx int, attributes attributeCollection) error {
	var (
		err         error
		imgRenderer = newTextOnImageRenderer()
	)

	// Initialize background
	if attributes.BackgroundColor != nil {
		if len(attributes.BackgroundColor) != 4 {
			return errors.New("Background color definition needs 4 hex values")
		}

		if err := ctx.Err(); err != nil {
			// Page context was cancelled, do not draw
			return err
		}

		imgRenderer.DrawBackgroundColor(attributes.BackgroundToColor())
	}

	if attributes.Image != "" {
		if err = imgRenderer.DrawBackgroundFromFile(attributes.Image); err != nil {
			return errors.Wrap(err, "Unable to draw background from disk")
		}
	}

	// Initialize color
	var textColor color.Color = color.RGBA{0xff, 0xff, 0xff, 0xff}
	if attributes.RGBA != nil {
		if len(attributes.RGBA) != 4 {
			return errors.New("RGBA color definition needs 4 hex values")
		}

		textColor = attributes.RGBAToColor()
	}

	// Initialize fontsize
	var fontsize float64 = 120
	if attributes.FontSize != nil {
		fontsize = *attributes.FontSize
	}

	border := 10
	if attributes.Border != nil {
		border = *attributes.Border
	}

	if strings.TrimSpace(attributes.Text) != "" {
		text := strings.TrimSpace(attributes.Text)
		if attributes.TextWithoutTrim {
			text = attributes.Text
		}
		if err = imgRenderer.DrawBigText(text, fontsize, border, textColor); err != nil {
			return errors.Wrap(err, "Unable to render text")
		}
	}

	if strings.TrimSpace(attributes.Caption) != "" {
		caption := strings.TrimSpace(attributes.Caption)
		if attributes.CaptionWithoutTrim {
			caption = attributes.Caption
		}
		if err = imgRenderer.DrawCaptionText(caption); err != nil {
			return errors.Wrap(err, "Unable to render caption")
		}
	}

	if !d.running && d.NeedsLoop(attributes) {
		return nil
	}

	if err := ctx.Err(); err != nil {
		// Page context was cancelled, do not draw
		return err
	}

	return errors.Wrap(sd.FillImage(idx, imgRenderer.GetImage()), "Unable to set image")
}

func (d displayElementText) NeedsLoop(attributes attributeCollection) bool {
	if attributes.BackgroundColor == nil {
		return false
	}

	if attributes.Interval > 0 && attributes.Interval < 100*time.Millisecond {
		log.WithFields(log.Fields{
			"type":     "text",
			"interval": attributes.Interval,
		}).Warn("Ignoring interval below 100ms")
		return false
	}

	return attributes.Interval >= 100*time.Millisecond
}

func (d *displayElementText) StartLoopDisplay(ctx context.Context, idx int, attributes attributeCollection) error {
	d.running = true

	withBg := attributes
	withoutBg := attributes
	withoutBg.BackgroundColor = nil

	go func() {
		showBackground := true
		for tick := time.NewTicker(attributes.Interval); ; <-tick.C {
			if ctx.Err() != nil || !d.running {
				return
			}

			renderAttrs := withBg
			if !showBackground {
				renderAttrs = withoutBg
			}

			if err := d.Display(ctx, idx, renderAttrs); err != nil {
				log.WithError(err).Error("Unable to refresh element")
			}

			showBackground = !showBackground
		}
	}()

	return nil
}

func (d *displayElementText) StopLoopDisplay() error {
	d.running = false
	return nil
}
