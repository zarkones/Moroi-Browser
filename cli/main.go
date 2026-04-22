package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"browser"
)

func main() {
	headless := flag.Bool("headless", true, "Start browser in headless mode (no visible window)")
	flag.Parse()

	fmt.Println("Starting Browser CLI tool...")
	fmt.Printf("Headless mode: %v\n\n", *headless)

	b, err := browser.New(*headless)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create browser: %v\n", err)
		os.Exit(1)
	}
	defer b.Close()

	fmt.Println("Browser ready! Type 'help' for commands.")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("browser> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("\nExiting...")
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]
		args := parts[1:]

		switch cmd {
		case "help":
			printHelp()

		case "navigate":
			if len(args) == 0 {
				fmt.Println("Usage: navigate <url>")
				continue
			}
			url := args[0]
			if err := b.Navigate(url); err != nil {
				fmt.Printf("Navigate failed: %v\n", err)
			} else {
				fmt.Printf("Navigated to %s\n", url)
			}

		case "snapshot":
			snap, err := b.Snapshot()
			if err != nil {
				fmt.Printf("Snapshot failed: %v\n", err)
			} else {
				fmt.Println(snap)
			}

		case "click":
			if len(args) == 0 {
				fmt.Println("Usage: click <ref>")
				continue
			}
			ref := args[0]
			if err := b.ClickByRef(ref); err != nil {
				fmt.Printf("Click failed: %v\n", err)
			} else {
				fmt.Printf("Clicked ref=%s\n", ref)
			}

		case "type":
			if len(args) < 2 {
				fmt.Println("Usage: type <ref> <text...>")
				continue
			}
			ref := args[0]
			text := strings.Join(args[1:], " ")
			if err := b.TypeByRef(ref, text); err != nil {
				fmt.Printf("Type failed: %v\n", err)
			} else {
				fmt.Printf("Typed into ref=%s\n", ref)
			}

		case "scroll":
			if len(args) == 0 {
				fmt.Println("Usage: scroll <ref>")
				continue
			}
			ref := args[0]
			if err := b.ScrollIntoViewByRef(ref); err != nil {
				fmt.Printf("Scroll failed: %v\n", err)
			} else {
				fmt.Printf("Scrolled ref=%s into view\n", ref)
			}

		case "url":
			u, err := b.GetCurrentURL()
			if err != nil {
				fmt.Printf("GetCurrentURL failed: %v\n", err)
			} else {
				fmt.Println(u)
			}

		case "html":
			h, err := b.GetHTML()
			if err != nil {
				fmt.Printf("GetHTML failed: %v\n", err)
			} else {
				fmt.Println(h)
			}

		case "text":
			t, err := b.GetText()
			if err != nil {
				fmt.Printf("GetText failed: %v\n", err)
			} else {
				fmt.Println(t)
			}

		case "back":
			if err := b.NavigateBack(); err != nil {
				fmt.Printf("NavigateBack failed: %v\n", err)
			} else {
				fmt.Println("Navigated back")
			}

		case "forward":
			if err := b.NavigateForward(); err != nil {
				fmt.Printf("NavigateForward failed: %v\n", err)
			} else {
				fmt.Println("Navigated forward")
			}

		case "refresh":
			if err := b.Refresh(); err != nil {
				fmt.Printf("Refresh failed: %v\n", err)
			} else {
				fmt.Println("Page refreshed")
			}

		case "screenshot":
			if len(args) == 0 {
				fmt.Println("Usage: screenshot <path>")
				continue
			}
			path := args[0]
			if err := b.TakeScreenshot(path); err != nil {
				fmt.Printf("Screenshot failed: %v\n", err)
			} else {
				fmt.Printf("Screenshot saved to %s\n", path)
			}

		case "quit", "exit", "close":
			fmt.Println("Closing browser...")
			b.Close()
			os.Exit(0)

		default:
			fmt.Printf("Unknown command: %s\nType 'help' for available commands.\n", cmd)
		}
	}
}

func printHelp() {
	fmt.Println(`
Browser CLI - controls all functions from the provided browser library

Commands:
  navigate <url>          Navigate to a URL
  snapshot                Take AI-readable hierarchical snapshot (updates internal refs)
  click <ref>             Click element using ref from last snapshot
  type <ref> <text...>    Type text into input/textarea by ref
  scroll <ref>            Scroll element into view by ref
  url                     Print current page URL
  html                    Print full page HTML
  text                    Print visible page text
  back                    Navigate back in history
  forward                 Navigate forward in history
  refresh                 Reload current page
  screenshot <path>       Capture screenshot and save to file (e.g. screenshot.png)
  help                    Show this help
  quit / exit / close     Close browser and exit CLI

Notes:
- Always run "snapshot" after navigation or page changes to get fresh refs.
- Refs (e.g. e1, e42) come from the last snapshot output.
- Text in "type" can contain spaces.
- The underlying browser instance stays alive for the whole session.`)
}
