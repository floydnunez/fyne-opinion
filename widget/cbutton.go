package widget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	col "fyne.io/fyne/v2/internal/color"
	"fyne.io/fyne/v2/internal/widget"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
)

var _ fyne.Focusable = (*CButton)(nil)

// Button widget has a text label and triggers an event func when clicked
type CButton struct { //Custom Buttton but that's a lot to types
	DisableableWidget
	Text string
	Icon fyne.Resource
	// Specify how prominent the button should be, High will highlight the button and Low will remove some decoration.
	//
	// Since: 1.4
	Importance    Importance
	Alignment     ButtonAlign
	IconPlacement ButtonIconPlacement

	OnTapped func() `json:"-"`

	hovered, focused bool
	tapAnim          *fyne.Animation
	background       *canvas.Rectangle
}

// NewButton creates a new button widget with the set label and tap handler
func NewCButton(label string, tapped func()) *CButton {
	button := &CButton{
		Text:     label,
		OnTapped: tapped,
	}

	button.ExtendBaseWidget(button)
	return button
}

// NewButtonWithIcon creates a new button widget with the specified label, themed icon and tap handler
func NewCButtonWithIcon(label string, icon fyne.Resource, tapped func()) *CButton {
	button := &CButton{
		Text:     label,
		Icon:     icon,
		OnTapped: tapped,
	}

	button.ExtendBaseWidget(button)
	return button
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (b *CButton) CreateRenderer() fyne.WidgetRenderer {
	b.ExtendBaseWidget(b)
	seg := &TextSegment{Text: b.Text, Style: RichTextStyleStrong}
	seg.Style.Alignment = fyne.TextAlignCenter
	text := NewRichText(seg)
	text.inset = fyne.NewSquareSize(theme.InnerPadding())

	b.background = canvas.NewRectangle(theme.ButtonColor())
	b.background.CornerRadius = theme.InputRadiusSize()
	tapBG := canvas.NewRectangle(color.Transparent)
	b.tapAnim = newButtonTapAnimation(tapBG, b)
	b.tapAnim.Curve = fyne.AnimationEaseOut
	objects := []fyne.CanvasObject{
		b.background,
		tapBG,
		text,
	}
	r := &cbuttonRenderer{
		BaseRenderer: widget.NewBaseRenderer(objects),
		background:   b.background,
		tapBG:        tapBG,
		button:       b,
		label:        text,
		layout:       layout.NewHBoxLayout(),
	}
	r.updateIconAndText()
	r.applyTheme()
	return r
}

// Cursor returns the cursor type of this widget
func (b *CButton) Cursor() desktop.Cursor {
	return desktop.DefaultCursor
}

// FocusGained is a hook called by the focus handling logic after this object gained the focus.
func (b *CButton) FocusGained() {
	b.focused = true
	b.Refresh()
}

// FocusLost is a hook called by the focus handling logic after this object lost the focus.
func (b *CButton) FocusLost() {
	b.focused = false
	b.Refresh()
}

// MinSize returns the size that this widget should not shrink below
func (b *CButton) MinSize() fyne.Size {
	b.ExtendBaseWidget(b)
	return b.BaseWidget.MinSize()
}

// MouseIn is called when a desktop pointer enters the widget
func (b *CButton) MouseIn(*desktop.MouseEvent) {
	b.hovered = true

	b.applyButtonTheme()
}

// MouseMoved is called when a desktop pointer hovers over the widget
func (b *CButton) MouseMoved(*desktop.MouseEvent) {
}

// MouseOut is called when a desktop pointer exits the widget
func (b *CButton) MouseOut() {
	b.hovered = false

	b.applyButtonTheme()
}

// SetIcon updates the icon on a label - pass nil to hide an icon
func (b *CButton) SetIcon(icon fyne.Resource) {
	b.Icon = icon

	b.Refresh()
}

// SetText allows the button label to be changed
func (b *CButton) SetText(text string) {
	b.Text = text

	b.Refresh()
}

// Tapped is called when a pointer tapped event is captured and triggers any tap handler
func (b *CButton) Tapped(*fyne.PointEvent) {
	if b.Disabled() {
		return
	}

	b.tapAnimation()
	b.applyButtonTheme()

	if b.OnTapped != nil {
		b.OnTapped()
	}
}

// TypedRune is a hook called by the input handling logic on text input events if this object is focused.
func (b *CButton) TypedRune(rune) {
}

// TypedKey is a hook called by the input handling logic on key events if this object is focused.
func (b *CButton) TypedKey(ev *fyne.KeyEvent) {
	if ev.Name == fyne.KeySpace {
		b.Tapped(nil)
	}
}

func (b *CButton) applyButtonTheme() {
	if b.background == nil {
		return
	}

	b.background.FillColor = b.buttonColor()
	b.background.CornerRadius = theme.InputRadiusSize()
	b.background.Refresh()
}

func (b *CButton) buttonColor() color.Color {
	switch {
	case b.Disabled():
		if b.Importance == LowImportance {
			return color.Transparent
		}
		return theme.DisabledButtonColor()
	case b.focused:
		bg := theme.ButtonColor()
		if b.Importance == HighImportance {
			bg = theme.PrimaryColor()
		} else if b.Importance == DangerImportance {
			bg = theme.ErrorColor()
		} else if b.Importance == WarningImportance {
			bg = theme.WarningColor()
		} else if b.Importance == SuccessImportance {
			bg = theme.SuccessColor()
		}

		return blendColor(bg, theme.FocusColor())
	case b.hovered:
		bg := theme.ButtonColor()
		if b.Importance == HighImportance {
			bg = theme.PrimaryColor()
		} else if b.Importance == DangerImportance {
			bg = theme.ErrorColor()
		} else if b.Importance == WarningImportance {
			bg = theme.WarningColor()
		} else if b.Importance == SuccessImportance {
			bg = theme.SuccessColor()
		}

		return blendColor(bg, theme.HoverColor())
	case b.Importance == HighImportance:
		return theme.PrimaryColor()
	case b.Importance == LowImportance:
		return color.Transparent
	case b.Importance == DangerImportance:
		return theme.ErrorColor()
	case b.Importance == WarningImportance:
		return theme.WarningColor()
	case b.Importance == SuccessImportance:
		return theme.SuccessColor()
	default:
		return theme.ButtonColor()
	}
}

func (b *CButton) tapAnimation() {
	if b.tapAnim == nil {
		return
	}
	b.tapAnim.Stop()

	if fyne.CurrentApp().Settings().ShowAnimations() {
		b.tapAnim.Start()
	}
}

type cbuttonRenderer struct {
	widget.BaseRenderer

	icon       *canvas.Image
	label      *RichText
	background *canvas.Rectangle
	tapBG      *canvas.Rectangle
	button     *CButton
	layout     fyne.Layout
}

// Layout the components of the button widget
func (r *cbuttonRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)
	r.tapBG.Resize(size)

	hasIcon := r.icon != nil
	hasLabel := r.label.Segments[0].(*TextSegment).Text != ""
	if !hasIcon && !hasLabel {
		// Nothing to layout
		return
	}
	iconSize := fyne.NewSquareSize(theme.IconInlineSize())
	labelSize := r.label.MinSize()
	padding := r.padding()
	if hasLabel {
		if hasIcon {
			// Both
			var objects []fyne.CanvasObject
			if r.button.IconPlacement == ButtonIconLeadingText {
				objects = append(objects, r.icon, r.label)
			} else {
				objects = append(objects, r.label, r.icon)
			}
			r.icon.SetMinSize(iconSize)
			min := r.layout.MinSize(objects)
			r.layout.Layout(objects, min)
			pos := alignedPosition(r.button.Alignment, padding, min, size)
			labelOff := (min.Height - labelSize.Height) / 2
			r.label.Move(r.label.Position().Add(pos).AddXY(0, labelOff))
			r.icon.Move(r.icon.Position().Add(pos))
		} else {
			// Label Only
			r.label.Move(alignedPosition(r.button.Alignment, padding, labelSize, size))
			r.label.Resize(labelSize)
		}
	} else {
		// Icon Only
		r.icon.Move(alignedPosition(r.button.Alignment, padding, iconSize, size))
		r.icon.Resize(iconSize)
	}
}

// MinSize calculates the minimum size of a button.
// This is based on the contained text, any icon that is set and a standard
// amount of padding added.
func (r *cbuttonRenderer) MinSize() (size fyne.Size) {
	hasIcon := r.icon != nil
	hasLabel := r.label.Segments[0].(*TextSegment).Text != ""
	iconSize := fyne.NewSquareSize(theme.IconInlineSize())
	labelSize := r.label.MinSize()
	if hasLabel {
		size.Width = labelSize.Width
	}
	if hasIcon {
		if hasLabel {
			size.Width += theme.Padding()
		}
		size.Width += iconSize.Width
	}
	size.Height = fyne.Max(labelSize.Height, iconSize.Height)
	size = size.Add(r.padding())
	return
}

func (r *cbuttonRenderer) Refresh() {
	r.label.inset = fyne.NewSize(theme.InnerPadding(), theme.InnerPadding())
	r.label.Segments[0].(*TextSegment).Text = r.button.Text
	r.updateIconAndText()
	r.applyTheme()
	r.background.Refresh()
	r.Layout(r.button.Size())
	canvas.Refresh(r.button.super())
}

// applyTheme updates this button to match the current theme
func (r *cbuttonRenderer) applyTheme() {
	r.button.applyButtonTheme()
	r.label.Segments[0].(*TextSegment).Style.ColorName = theme.ColorNameForeground
	switch {
	case r.button.disabled:
		r.label.Segments[0].(*TextSegment).Style.ColorName = theme.ColorNameDisabled
	case r.button.Importance == HighImportance || r.button.Importance == DangerImportance || r.button.Importance == WarningImportance || r.button.Importance == SuccessImportance:
		if r.button.focused {
			r.label.Segments[0].(*TextSegment).Style.ColorName = theme.ColorNameForeground
		} else {
			r.label.Segments[0].(*TextSegment).Style.ColorName = theme.ColorNameBackground
		}
	}
	r.label.Refresh()
	if r.icon != nil && r.icon.Resource != nil {
		switch res := r.icon.Resource.(type) {
		case *theme.ThemedResource:
			if r.button.Importance == HighImportance || r.button.Importance == DangerImportance || r.button.Importance == WarningImportance || r.button.Importance == SuccessImportance {
				r.icon.Resource = theme.NewInvertedThemedResource(res)
				r.icon.Refresh()
			}
		case *theme.InvertedThemedResource:
			if r.button.Importance != HighImportance && r.button.Importance != DangerImportance && r.button.Importance != WarningImportance && r.button.Importance != SuccessImportance {
				r.icon.Resource = res.Original()
				r.icon.Refresh()
			}
		}
	}
}

func (r *cbuttonRenderer) padding() fyne.Size {
	return fyne.NewSquareSize(theme.InnerPadding() * 2)
}

func (r *cbuttonRenderer) updateIconAndText() {
	if r.button.Icon != nil && r.button.Visible() {
		if r.icon == nil {
			r.icon = canvas.NewImageFromResource(r.button.Icon)
			r.icon.FillMode = canvas.ImageFillContain
			r.SetObjects([]fyne.CanvasObject{r.background, r.tapBG, r.label, r.icon})
		}
		if r.button.Disabled() {
			r.icon.Resource = theme.NewDisabledResource(r.button.Icon)
		} else {
			r.icon.Resource = r.button.Icon
		}
		r.icon.Refresh()
		r.icon.Show()
	} else if r.icon != nil {
		r.icon.Hide()
	}
	if r.button.Text == "" {
		r.label.Hide()
	} else {
		r.label.Show()
	}
	r.label.Refresh()
}

func newCButtonTapAnimation(bg *canvas.Rectangle, w fyne.Widget) *fyne.Animation {
	return fyne.NewAnimation(canvas.DurationStandard, func(done float32) {
		mid := w.Size().Width / 2
		size := mid * done
		bg.Resize(fyne.NewSize(size*2, w.Size().Height))
		bg.Move(fyne.NewPos(mid-size, 0))

		r, g, bb, a := col.ToNRGBA(theme.PressedColor())
		aa := uint8(a)
		fade := aa - uint8(float32(aa)*done)
		if fade > 0 {
			bg.FillColor = &color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(bb), A: fade}
		} else {
			bg.FillColor = color.Transparent
		}
		canvas.Refresh(bg)
	})
}
