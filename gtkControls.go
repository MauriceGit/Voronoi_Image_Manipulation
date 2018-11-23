// gtkControls.go
package main

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/andlabs/ui"
	_ "github.com/andlabs/ui/winmanifest"
	"github.com/bamiaux/rez"
)

var (
	functionChannel   chan func()
	imageHandler      *imageTableHandler
	tableModelHandler *ui.TableModel
	imageFilename     string
	vLineColor        = [...]float64{1, 1, 1, 1}
	dLineColor        = [...]float64{1, 1, 1, 1}
	pointColor        = [...]float64{1, 1, 1, 1}
	chColor           = [...]float64{1, 1, 1, 1}
)

type imageTableHandler struct {
	selectedImage *ui.Image
}

func loadImageRGBA(filename string) (*image.RGBA, float32) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	img, _, err := image.Decode(file)

	if err != nil {
		fmt.Println(err)
	}

	b := img.Bounds()
	w := b.Max.X - b.Min.X
	h := b.Max.Y - b.Min.Y
	ratio := float32(w) / float32(h)
	scaledImage := image.NewRGBA(image.Rect(b.Min.X, b.Min.Y, int(150*ratio)+b.Min.X, 150+b.Min.Y))

	err = rez.Convert(scaledImage, img, rez.NewBicubicFilter())
	if err != nil {
		fmt.Println(err)
	}
	return scaledImage, ratio
}

func createUIImage(filename string) *ui.Image {
	scaledImage, ratio := loadImageRGBA(filename)

	uiImg := ui.NewImage(150*float64(ratio), 150)
	uiImg.Append(scaledImage)
	return uiImg
}

func newModelHandler(filename string) *imageTableHandler {
	m := new(imageTableHandler)
	m.selectedImage = createUIImage(filename)
	return m
}
func (mh *imageTableHandler) ColumnTypes(m *ui.TableModel) []ui.TableValue {
	return []ui.TableValue{
		ui.TableImage{},
		// Color not used right now.
		ui.TableColor{},
	}
}
func (mh *imageTableHandler) NumRows(m *ui.TableModel) int {
	return 1
}
func (mh *imageTableHandler) CellValue(m *ui.TableModel, row, column int) ui.TableValue {

	if row == 0 && column == 1 {
		//return ui.TableColor{0.5, 0.5, 0, 1}
	}

	return ui.TableImage{mh.selectedImage}
}
func (mh *imageTableHandler) SetCellValue(m *ui.TableModel, row, column int, value ui.TableValue) {
	fmt.Printf("Set Cell Value\n")
}

func createFileOpenButton(mainwin *ui.Window, c chan func(), handler *imageTableHandler, tableHandler *ui.TableModel) *ui.Button {
	button := ui.NewButton("Open New Image")
	button.OnClicked(func(*ui.Button) {
		filename := ui.OpenFile(mainwin)
		if filename != "" {
			c <- func() {
				SetNewImage(filename)
				ReadyForRebuild(true)
				// It seems like there is an issue with data not being fully uploaded to GPU before we try to render them.
				// By giving it a bit more time, the rendering works more reliably.
				time.Sleep(50 * time.Millisecond)
				ReadyForRender(true)
			}
			handler.selectedImage = createUIImage(filename)
			tableHandler.RowChanged(0)
		}
	})
	return button
}

func createFileSaveButton(mainwin *ui.Window, c chan func()) *ui.Button {
	button := ui.NewButton("Save Image")
	button.OnClicked(func(*ui.Button) {
		filename := ui.SaveFile(mainwin)
		if filename != "" {

			if !strings.HasSuffix(filename, ".png") {
				filename = filename + ".png"
			}

			c <- func() {
				SaveImage(filename)
			}
		}
	})
	return button
}

func createImageOperations(mainwin *ui.Window, c chan func(), handler *imageTableHandler, tableHandler *ui.TableModel) *ui.Grid {
	grid := ui.NewGrid()
	grid.SetPadded(true)

	imageLoad := createFileOpenButton(mainwin, c, handler, tableHandler)
	imageSave := createFileSaveButton(mainwin, c)

	grid.Append(imageLoad, 0, 0, 1, 1, false, ui.AlignFill, true, ui.AlignFill)
	grid.Append(imageSave, 0, 1, 1, 1, false, ui.AlignFill, true, ui.AlignFill)

	return grid
}

func createImageTable(filename string) (*ui.Box, *imageTableHandler, *ui.TableModel) {
	modelHandler := newModelHandler(filename)
	tableModel := ui.NewTableModel(modelHandler)
	table := ui.NewTable(&ui.TableParams{
		Model:                         tableModel,
		RowBackgroundColorModelColumn: 1,
	})

	// The integer (0) is an index into the []ui.TableValue{} slice to determine the type at runtime!
	table.AppendImageColumn("Selected Image (preview)", 0)

	table.Disable()

	hbox := ui.NewVerticalBox()
	hbox.SetPadded(false)
	hbox.Append(table, true)
	return hbox, modelHandler, tableModel
}

func createPointCountButtons(c chan func()) *ui.Box {
	hbox := ui.NewHorizontalBox()

	decrease := ui.NewButton("-")
	increase := ui.NewButton("+")

	hbox.Append(decrease, false)
	hbox.Append(increase, false)

	decrease.OnClicked(func(*ui.Button) {
		c <- func() {
			DecreasePointCount()
			ReadyForRebuild(true)
			ReadyForRender(true)
		}
	})

	increase.OnClicked(func(*ui.Button) {
		c <- func() {
			IncreasePointCount()
			ReadyForRebuild(true)
			ReadyForRender(true)
		}
	})

	return hbox
}

func createPointDistributionButtons(c chan func()) *ui.RadioButtons {
	rb := ui.NewRadioButtons()
	rb.Append("Poisson Disk")
	rb.Append("Random")
	rb.Append("Grid")

	rb.SetSelected(0)

	rb.OnSelected(func(*ui.RadioButtons) {

		switch rb.Selected() {
		case 0:
			c <- func() {
				SetPointDistributionMethod(POINT_DISTRIBUTION_POISSON)
			}
		case 1:
			c <- func() {
				SetPointDistributionMethod(POINT_DISTRIBUTION_RANDOM)
			}
		case 2:
			c <- func() {
				SetPointDistributionMethod(POINT_DISTRIBUTION_GRID)
			}
		}

		c <- func() {
			ReadyForRebuild(true)
			ReadyForRender(true)
		}
	})

	return rb
}

func createFaceRenderingButtons(c chan func()) *ui.RadioButtons {
	rb := ui.NewRadioButtons()
	rb.Append("Delaunay Triangles")
	rb.Append("Voronoi Cells")
	rb.Append("Nothing")

	rb.SetSelected(1)

	rb.OnSelected(func(*ui.RadioButtons) {
		selectedIndex := rb.Selected()
		c <- func() {
			SetRenderTriangles(selectedIndex == 0)
		}
		c <- func() {
			SetRenderVoronoiCells(selectedIndex == 1)
		}
		// We re-render everything no matter what happened after the user selected the radio button.
		c <- func() {
			ReadyForRender(true)
		}

	})

	return rb
}

func createGeneralCheckboxes(c chan func()) *ui.Grid {
	grid := ui.NewGrid()
	grid.SetPadded(true)

	ve := ui.NewCheckbox("Voronoi Edges")
	de := ui.NewCheckbox("Delaunay Edges")
	p := ui.NewCheckbox("Points")
	ch := ui.NewCheckbox("Convex Hull")
	fc := ui.NewCheckbox("Adaptive Color")

	grid.Append(ve, 0, 0, 1, 1, false, ui.AlignFill, true, ui.AlignFill)
	grid.Append(de, 1, 0, 1, 1, false, ui.AlignFill, true, ui.AlignFill)
	grid.Append(p, 0, 1, 1, 1, false, ui.AlignFill, true, ui.AlignFill)
	grid.Append(ch, 1, 1, 1, 1, false, ui.AlignFill, true, ui.AlignFill)
	grid.Append(fc, 0, 2, 1, 1, false, ui.AlignFill, true, ui.AlignFill)

	ve.SetChecked(false)
	de.SetChecked(false)
	p.SetChecked(false)
	ch.SetChecked(false)
	fc.SetChecked(false)

	ve.OnToggled(func(*ui.Checkbox) {
		c <- func() {
			SetRenderVoronoiEdges(ve.Checked())
			ReadyForRender(true)
		}
	})
	de.OnToggled(func(*ui.Checkbox) {
		c <- func() {
			SetRenderLines(de.Checked())
			ReadyForRender(true)
		}
	})
	p.OnToggled(func(*ui.Checkbox) {
		c <- func() {
			SetRenderPoints(p.Checked())
			ReadyForRender(true)
		}
	})
	ch.OnToggled(func(*ui.Checkbox) {
		c <- func() {
			SetRenderConvexHull(ch.Checked())
			ReadyForRender(true)
		}
	})
	fc.OnToggled(func(*ui.Checkbox) {
		c <- func() {
			SetUseExternalColor(!fc.Checked())
			ReadyForRender(true)
		}
	})

	return grid
}

func createVoronoiColorButton(c chan func()) *ui.ColorButton {
	b := ui.NewColorButton()
	b.SetColor(vLineColor[0], vLineColor[1], vLineColor[2], vLineColor[3])

	b.OnChanged(func(*ui.ColorButton) {
		c <- func() {
			SetVoronoiLineColor(b.Color())
			ReadyForRender(true)
		}
	})

	return b
}

func createDelaunayColorButton(c chan func()) *ui.ColorButton {
	b := ui.NewColorButton()
	b.SetColor(dLineColor[0], dLineColor[1], dLineColor[2], dLineColor[3])

	b.OnChanged(func(*ui.ColorButton) {
		c <- func() {
			SetDelaunayLineColor(b.Color())
			ReadyForRender(true)
		}
	})

	return b
}

func createPointColorButton(c chan func()) *ui.ColorButton {
	b := ui.NewColorButton()
	b.SetColor(pointColor[0], pointColor[1], pointColor[2], pointColor[3])

	b.OnChanged(func(*ui.ColorButton) {
		c <- func() {
			SetPointColor(b.Color())
			ReadyForRender(true)
		}
	})

	return b
}

func createCHColorButton(c chan func()) *ui.ColorButton {
	b := ui.NewColorButton()
	b.SetColor(chColor[0], chColor[1], chColor[2], chColor[3])

	b.OnChanged(func(*ui.ColorButton) {
		c <- func() {
			SetCHColor(b.Color())
			ReadyForRender(true)
		}
	})

	return b
}

func setupUI() {

	mainwin := ui.NewWindow("Geometry Controls", 360, 680, true)
	mainwin.SetMargined(true)
	mainwin.OnClosing(func(*ui.Window) bool {
		fmt.Printf("Close.\n")
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		fmt.Printf("Destroy.\n")
		mainwin.Destroy()
		return true
	})

	grid := ui.NewGrid()
	grid.SetPadded(true)

	image, handler, tableModel := createImageTable(imageFilename)
	imageHandler = handler
	tableModelHandler = tableModel
	//imageSelect := createFileOpenButton(mainwin, functionChannel, imageHandler, tableModelHandler)
	//imageSave := createFileSaveButton(mainwin, functionChannel)

	imageGrid := createImageOperations(mainwin, functionChannel, imageHandler, tableModelHandler)

	pointLable := ui.NewLabel("Point Count")
	pointButtons := createPointCountButtons(functionChannel)

	distLable := ui.NewLabel("Point Distribution")
	distButton := createPointDistributionButtons(functionChannel)

	faceLable := ui.NewLabel("Face rendering")
	faceButton := createFaceRenderingButtons(functionChannel)

	generalLable := ui.NewLabel("General")
	generalBoxes := createGeneralCheckboxes(functionChannel)

	dColorLable := ui.NewLabel("Voronoi Color")
	dColorButton := createVoronoiColorButton(functionChannel)

	vColorLable := ui.NewLabel("Delaunay Color")
	vColorButton := createDelaunayColorButton(functionChannel)

	pColorLable := ui.NewLabel("Point Color")
	pColorButton := createPointColorButton(functionChannel)

	chColorLable := ui.NewLabel("Convex Hull Color")
	chColorButton := createCHColorButton(functionChannel)

	gridYPos := 0
	grid.Append(imageGrid, 0, gridYPos, 1, 1, false, ui.AlignFill, false, ui.AlignStart)
	grid.Append(image, 1, gridYPos, 1, 1, true, ui.AlignFill, true, ui.AlignFill)
	gridYPos++
	grid.Append(ui.NewHorizontalSeparator(), 0, gridYPos, 2, 1, false, ui.AlignFill, false, ui.AlignFill)
	gridYPos++

	grid.Append(pointLable, 0, gridYPos, 1, 1, false, ui.AlignFill, false, ui.AlignFill)
	grid.Append(pointButtons, 1, gridYPos, 1, 1, false, ui.AlignFill, false, ui.AlignFill)
	gridYPos++
	grid.Append(ui.NewHorizontalSeparator(), 0, gridYPos, 2, 1, false, ui.AlignFill, false, ui.AlignFill)
	gridYPos++

	grid.Append(distLable, 0, gridYPos, 1, 1, false, ui.AlignFill, false, ui.AlignFill)
	grid.Append(distButton, 1, gridYPos, 1, 1, false, ui.AlignFill, false, ui.AlignFill)
	gridYPos++
	grid.Append(ui.NewHorizontalSeparator(), 0, gridYPos, 2, 1, false, ui.AlignFill, false, ui.AlignFill)
	gridYPos++

	grid.Append(faceLable, 0, gridYPos, 1, 1, false, ui.AlignFill, false, ui.AlignFill)
	grid.Append(faceButton, 1, gridYPos, 1, 1, false, ui.AlignFill, false, ui.AlignFill)
	gridYPos++
	grid.Append(ui.NewHorizontalSeparator(), 0, gridYPos, 2, 1, false, ui.AlignFill, false, ui.AlignFill)
	gridYPos++

	grid.Append(generalLable, 0, gridYPos, 1, 1, false, ui.AlignFill, false, ui.AlignFill)
	grid.Append(generalBoxes, 1, gridYPos, 1, 1, false, ui.AlignFill, false, ui.AlignFill)
	gridYPos++
	grid.Append(ui.NewHorizontalSeparator(), 0, gridYPos, 2, 1, false, ui.AlignFill, false, ui.AlignFill)
	gridYPos++

	grid.Append(dColorLable, 0, gridYPos, 1, 1, false, ui.AlignFill, false, ui.AlignFill)
	grid.Append(dColorButton, 1, gridYPos, 1, 1, false, ui.AlignStart, false, ui.AlignFill)
	gridYPos++
	grid.Append(vColorLable, 0, gridYPos, 1, 1, false, ui.AlignFill, false, ui.AlignFill)
	grid.Append(vColorButton, 1, gridYPos, 1, 1, false, ui.AlignStart, false, ui.AlignFill)
	gridYPos++
	grid.Append(pColorLable, 0, gridYPos, 1, 1, false, ui.AlignFill, false, ui.AlignFill)
	grid.Append(pColorButton, 1, gridYPos, 1, 1, false, ui.AlignStart, false, ui.AlignFill)
	gridYPos++
	grid.Append(chColorLable, 0, gridYPos, 1, 1, false, ui.AlignFill, false, ui.AlignFill)
	grid.Append(chColorButton, 1, gridYPos, 1, 1, false, ui.AlignStart, false, ui.AlignFill)
	gridYPos++

	mainwin.SetChild(grid)

	mainwin.Show()

}

// Sets the default rendering values according to the ui so we start in a consistent state.
// TODO: Move those to constants!
func setDefaultRenderValues(c chan func()) {
	// This is a blocking sent to the channel!!!
	// That means, it blocks until the main rendering thread is initialized and is able to pull from the channel.
	// This is OK because we are in the initialization phase anyway.
	c <- func() {
		SetPointDistributionMethod(POINT_DISTRIBUTION_POISSON)

		SetRenderTriangles(false)
		SetRenderVoronoiCells(true)

		SetRenderVoronoiEdges(false)
		SetRenderLines(false)
		SetRenderPoints(false)
		SetRenderConvexHull(false)
		SetUseExternalColor(true)

		SetVoronoiLineColor(vLineColor[0], vLineColor[1], vLineColor[2], vLineColor[3])
		SetDelaunayLineColor(dLineColor[0], dLineColor[1], dLineColor[2], dLineColor[3])
		SetPointColor(pointColor[0], pointColor[1], pointColor[2], pointColor[3])
		SetCHColor(chColor[0], chColor[1], chColor[2], chColor[3])

		ReadyForRebuild(true)
		ReadyForRender(true)
	}
}

func createGUI(wg *sync.WaitGroup, c chan func(), pointCount int, defaultImage string) {

	defer wg.Done()

	imageFilename = defaultImage
	functionChannel = c

	setDefaultRenderValues(c)

	ui.Main(setupUI)

	f := func() {
		CloseWindow()
	}

	// Non-Blocking write to channel in case it is already closed...
	select {
	case c <- f:
	default:
	}

}

func main() {

	functionChannel := make(chan func())
	var wg sync.WaitGroup
	wg.Add(1)

	closingChannel := make(chan int, 1)

	go func() {
		<-closingChannel
		ui.Quit()
	}()

	initialPointCount := 192
	dir, _ := filepath.Split(os.Args[0])
	defaultImage := dir + "Images/apple.png"

	go createGUI(&wg, functionChannel, initialPointCount, defaultImage)

	InitializeGeoRender(functionChannel, closingChannel, initialPointCount, defaultImage)

	wg.Wait()

	// In case we come here and the closing channel did not receive any close commands, we provide them manually so the go-routine can finish nicely.
	select {
	case closingChannel <- 1:
	default:
	}

}
