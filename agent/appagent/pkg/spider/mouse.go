package spider

import (
	"math"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func CircleMoveMouse(page *rod.Page) error {
	// Get page dimensions
	dimensions, err := page.Eval(`() => {
		return {
			width: window.innerWidth,
			height: window.innerHeight
		}
	}`)
	if err != nil {
		return err
	}

	width := dimensions.Value.Get("width").Int()
	height := dimensions.Value.Get("height").Int()

	// Calculate center and radius
	centerX := width / 2
	centerY := height / 2
	radius := int(math.Min(float64(width), float64(height)) * 0.3)

	// Move mouse in a circle
	steps := 12
	for i := 0; i < steps; i++ {
		angle := 2 * math.Pi * float64(i) / float64(steps)
		x := centerX + int(float64(radius)*math.Cos(angle))
		y := centerY + int(float64(radius)*math.Sin(angle))

		err := page.Mouse.MoveTo(proto.Point{X: float64(x), Y: float64(y)})
		if err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}
