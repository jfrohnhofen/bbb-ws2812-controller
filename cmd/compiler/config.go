package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const NumChannels = 48

type Point [2]float64

func (p Point) add(other Point) Point {
	return Point{p[0] + other[0], p[1] + other[1]}
}

func (p Point) X() float64 {
	return p[0]
}

func (p Point) Y() float64 {
	return p[1]
}

type Panel struct {
	Channel       uint
	TopLeft       Point
	BottomRight   Point
	Input         Point
	Leds          []Point
	NumBufferLeds uint
}

func (p Panel) NumLeds() uint {
	return uint(len(p.Leds)) + p.NumBufferLeds
}

type Config struct {
	ImageScale float64
	Panels     []Panel
}

type panelLayout struct {
	Size  Point
	Input Point
	Leds  []Point
}

type panel struct {
	Channel       uint
	Layout        string
	Pos           Point
	NumBufferLeds uint
}

type config struct {
	ImageScale   float64
	PanelLayouts map[string]panelLayout
	Panels       []panel
}

func ReadConfig(file string) (Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Config{}, err
	}

	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	channels := map[uint]Panel{}

	for _, panel := range cfg.Panels {
		layout, ok := cfg.PanelLayouts[panel.Layout]
		if !ok {
			return Config{}, fmt.Errorf("panel layout '%s' not found", panel.Layout)
		}

		if panel.Channel >= NumChannels {
			return Config{}, fmt.Errorf("channel %d outside of range", panel.Channel)
		}
		if _, ok := channels[panel.Channel]; ok {
			return Config{}, fmt.Errorf("channel %d already in use", panel.Channel)
		}

		leds := []Point{}
		for _, led := range layout.Leds {
			leds = append(leds, led.add(panel.Pos))
		}
		channels[panel.Channel] = Panel{
			Channel:       panel.Channel,
			TopLeft:       panel.Pos,
			BottomRight:   layout.Size.add(panel.Pos),
			Input:         layout.Input.add(panel.Pos),
			Leds:          leds,
			NumBufferLeds: panel.NumBufferLeds,
		}
	}

	panels := []Panel{}
	for i := uint(0); i < NumChannels; i++ {
		if panel, ok := channels[i]; ok {
			panels = append(panels, panel)
		}
	}

	return Config{
		ImageScale: cfg.ImageScale,
		Panels:     panels,
	}, nil
}
