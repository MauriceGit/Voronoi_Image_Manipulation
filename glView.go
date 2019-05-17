package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"runtime"
	"time"

	geo "github.com/MauriceGit/mtGeometry"
	mtgl "github.com/MauriceGit/mtOpenGL"
	v "github.com/MauriceGit/mtVector"
	sc "github.com/MauriceGit/sweepcircle"

	"github.com/fogleman/gg"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	g_windowTitle    = "Delaunay/Voronoi Image Manipulation"
	g_delaunayMargin = 10.0
	g_maxWindowSize  = 1000
)

const (
	POINT_DISTRIBUTION_RANDOM  = iota
	POINT_DISTRIBUTION_GRID    = iota
	POINT_DISTRIBUTION_POISSON = iota
)

///////////////////////////////////////////////////////
// FPS
///////////////////////////////////////////////////////
var g_timeSum float32 = 0.0
var g_currentTime float64 = 0.0
var g_lastCallTime float64 = 0.0
var g_frameCount int = 0
var g_fps float32 = 60.0

///////////////////////////////////////////////////////
// Delaunay
///////////////////////////////////////////////////////
var g_windowWidth int = 1000
var g_windowHeight int = 1000

//var g_windowWidth float64 = 1000
//var g_windowHeight float64 = 1000

var g_delaunayPointCount int
var g_delaunayTriangleGLBuffer geo.ArrayGeometry
var g_delaunayEdgesGLBuffer geo.ArrayGeometry
var g_delaunayPointsGLBuffer geo.Geometry
var g_convexHullGLBuffer geo.Geometry
var g_voronoiEdgesGLBuffer geo.ArrayGeometry
var g_voronoiTriangleGLBuffer geo.ArrayGeometry

///////////////////////////////////////////////////////
// Camera
///////////////////////////////////////////////////////
var g_fovy = mgl32.DegToRad(90.0)
var g_aspect = float32(g_windowWidth) / float32(g_windowHeight)
var g_nearPlane = float32(0.1)
var g_farPlane = float32(2000.0)

///////////////////////////////////////////////////////
// Delaunay Rendering Options
///////////////////////////////////////////////////////
var g_delaunayDistribution int = POINT_DISTRIBUTION_POISSON
var g_delaunayTexture mtgl.ImageTexture
var g_showDelaunayTexture = false
var g_renderVoronoiCells = false
var g_renderVoronoiEdges = false
var g_renderTriangles = false
var g_renderLines = false
var g_renderPoints = false
var g_renderConvexHull = false
var g_useExternalColor int32 = 0
var g_voronoiLineColor mgl32.Vec4
var g_delaunayLineColor mgl32.Vec4
var g_pointColor mgl32.Vec4
var g_chColor mgl32.Vec4

///////////////////////////////////////////////////////
// OpenGL Setup
///////////////////////////////////////////////////////
var g_delaunayTrianglesShader uint32
var g_delaunayEdgesShader uint32
var g_delaunayPointsShader uint32
var g_sceneColorTexMS uint32
var g_sceneDepthTexMS uint32
var g_sceneFboMS uint32

///////////////////////////////////////////////////////
// Control Communication
///////////////////////////////////////////////////////
var g_controlCommunication chan func()
var g_controlCommunicationClose chan int
var g_readyForRebuild bool = false
var g_readyForRender bool = false

///////////////////////////////////////////////////////
// GLFW and Other
///////////////////////////////////////////////////////
var g_window *glfw.Window

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func drawImage(d sc.Delaunay, name string) {
	var scale float64 = 1.0
	var imageSizeX float64 = 1000
	var imageSizeY float64 = 1000
	dc := gg.NewContext(int(imageSizeX), int(imageSizeY))

	// Background filling in white
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	dc.SetLineWidth(1.0)

	for i := 1; i < 10; i++ {

		x := float64(i) * 100 * scale
		y := imageSizeY - float64(i)*100*scale

		dc.SetRGB(1, 0.5, 0.5)
		dc.DrawLine(0, y, imageSizeX, y)
		dc.Stroke()
		dc.DrawLine(x, 0, x, imageSizeY)
		dc.Stroke()

		dc.SetRGB(1, 0.0, 0.0)
		// X axis
		//dc.DrawString(strconv.Itoa(int(x)), x+10, imageSizeY-10)
		//// Y axis
		//dc.DrawString(strconv.Itoa(int(imageSizeY-y)), 10, y-10)

	}

	dc.SetLineWidth(2.0)
	for i, e := range d.Edges {

		if e == sc.EmptyE {
			continue
		}

		dc.SetRGB(0, 0, 0)

		v1 := v.Vector{}
		v2 := v.Vector{}

		// For the Voronoi case
		shouldContinue := false
		switch {
		// Both are empty
		case e.VOrigin == sc.EmptyVertex && d.Edges[e.ETwin].VOrigin == sc.EmptyVertex:
			shouldContinue = true
		// Only the left one is empty
		case e.VOrigin == sc.EmptyVertex:
			v2 = d.Vertices[d.Edges[e.ETwin].VOrigin].Pos
			v1 = v.Add(e.TmpEdge.Pos, v.Mult(e.TmpEdge.Dir, 0.2))
		// Only the right one is empty
		case d.Edges[e.ETwin].VOrigin == sc.EmptyVertex:
			v1 = d.Vertices[e.VOrigin].Pos
			v2 = v.Add(e.TmpEdge.Pos, v.Mult(e.TmpEdge.Dir, 0.2))
		// No one is empty!
		default:
			v1 = d.Vertices[e.VOrigin].Pos
			v2 = d.Vertices[d.Edges[e.ETwin].VOrigin].Pos
		}

		if shouldContinue {
			continue
		}

		dc.DrawLine(v1.X*scale, imageSizeY-v1.Y*scale, v2.X*scale, imageSizeY-v2.Y*scale)
		dc.Stroke()

		dc.SetRGB(0, 0, 1)
		dc.DrawString(fmt.Sprintf("(%.1f, %.1f)", v1.X, v1.Y), v1.X, imageSizeY-v1.Y)

		dc.SetRGB(0, 0.5, 0)
		middleP := v.Vector{(v1.X + v2.X) / 2., (v1.Y + v2.Y) / 2., 0}
		ortho := v.Vector{0, 0, 1}
		crossP := v.Cross(v.Sub(v1, v2), ortho)
		crossP.Div(v.Length(crossP))
		crossP.Mult(15.)

		middleP.Add(crossP)

		i = i
		s := fmt.Sprintf("(%d)", i)
		dc.DrawStringAnchored(s, middleP.X, imageSizeY-middleP.Y, 0.5, 0.5)
	}

	dc.SetLineWidth(1.0)
	dc.SetRGB(1, 0, 0)
	for i, v := range d.Vertices {

		if v == sc.EmptyV {
			continue
		}

		dc.DrawCircle(v.Pos.X*scale, imageSizeY-v.Pos.Y*scale, 2)
		dc.Fill()

		i = i
		//s := fmt.Sprintf("(%d)", i)
		//dc.DrawStringAnchored(s, v.Pos.X-10, imageSizeY-v.Pos.Y-10, 0.5, 0.5)
	}

	//dc.SetRGB(1, 1, 0)
	//dc.DrawCircle(432, imageSizeY-894, 5)
	//dc.DrawCircle(599, imageSizeY-532, 5)
	//dc.DrawCircle(501, imageSizeY-578, 5)
	//dc.Fill()

	dc.SavePNG(name + ".png")
}

func createDelaunay(count int, rangeX, rangeY, margin float64) sc.Delaunay {

	var list v.PointList

	var seed int64 = int64(count)

	switch g_delaunayDistribution {
	case POINT_DISTRIBUTION_POISSON:

		adjustedCount := count
		if count < 3 {
			adjustedCount = 3
		}

		list = CreateFastPoissonDiscPoints(adjustedCount, rangeX, rangeY, margin, 30, seed)

	case POINT_DISTRIBUTION_RANDOM:
		list = CreateRandomPoints(count, rangeX, rangeY, margin, seed)

	case POINT_DISTRIBUTION_GRID:
		list = CreateShiftedGridPoints(count, rangeX, rangeY, margin)

	default:
		fmt.Println("No point distribution selected. Default to random.")
		list = CreateRandomPoints(count, rangeX, rangeY, margin, seed)
	}

	fmt.Printf("Points: %d\n", len(list))

	return sc.Triangulate(list)
}

func createDelaunayGLBuffer(d sc.Delaunay, rangeX, rangeY float64) geo.ArrayGeometry {
	mesh := make([]geo.Mesh, len(d.Faces)*3)

	for i, f := range d.Faces {
		v1 := d.Vertices[d.Edges[f.EEdge].VOrigin].Pos
		v2 := d.Vertices[d.Edges[d.Edges[f.EEdge].ENext].VOrigin].Pos
		v3 := d.Vertices[d.Edges[d.Edges[d.Edges[f.EEdge].ENext].ENext].VOrigin].Pos

		uv1 := mgl32.Vec2{float32(v1.X / rangeX), float32(v1.Y / rangeY)}
		uv2 := mgl32.Vec2{float32(v2.X / rangeX), float32(v2.Y / rangeY)}
		uv3 := mgl32.Vec2{float32(v3.X / rangeX), float32(v3.Y / rangeY)}

		averageUV := uv1.Add(uv2.Add(uv3)).Mul(1.0 / 3.0)

		mesh[i*3] = geo.Mesh{mgl32.Vec3{float32(v1.X), float32(v1.Y), float32(v1.Z)}, mgl32.Vec3{0.0, 0.0, 1.0}, averageUV}
		mesh[i*3+1] = geo.Mesh{mgl32.Vec3{float32(v2.X), float32(v2.Y), float32(v2.Z)}, mgl32.Vec3{0.0, 0.0, 1.0}, averageUV}
		mesh[i*3+2] = geo.Mesh{mgl32.Vec3{float32(v3.X), float32(v3.Y), float32(v3.Z)}, mgl32.Vec3{0.0, 0.0, 1.0}, averageUV}
	}

	return geo.GenerateGeometryArrayAttributes(&mesh, len(mesh))
}

func createVoronoiGLBuffer(vo sc.Voronoi, rangeX, rangeY float64) geo.ArrayGeometry {
	mesh := make([]geo.Mesh, 0)

	emptyF := sc.HEFace{}

	for _, f := range vo.Faces {

		if f == emptyF {
			break
		}

		e0 := f.EEdge
		e1 := vo.Edges[e0].ENext

		v0 := v.Vector{}
		v1 := v.Vector{}

		// For the Voronoi case
		shouldBreak := false
		switch {
		// Both are empty
		case vo.Edges[e0].VOrigin == sc.EmptyVertex && vo.Edges[vo.Edges[e0].ETwin].VOrigin == sc.EmptyVertex:
			shouldBreak = true
		// Only the left one is empty
		case vo.Edges[e0].VOrigin == sc.EmptyVertex:
			v1 = vo.Vertices[vo.Edges[vo.Edges[e0].ETwin].VOrigin].Pos
			v0 = v.Add(vo.Edges[e0].TmpEdge.Pos, v.Mult(vo.Edges[e0].TmpEdge.Dir, -10.0))
		// Only the right one is empty
		case vo.Edges[vo.Edges[e0].ETwin].VOrigin == sc.EmptyVertex:
			v0 = vo.Vertices[vo.Edges[e0].VOrigin].Pos
			v1 = v.Add(vo.Edges[e0].TmpEdge.Pos, v.Mult(vo.Edges[e0].TmpEdge.Dir, -10.0))
		// No one is empty!
		default:
			v0 = vo.Vertices[vo.Edges[e0].VOrigin].Pos
			v1 = vo.Vertices[vo.Edges[vo.Edges[e0].ETwin].VOrigin].Pos
		}

		if shouldBreak {
			break
		}

		p := vo.Faces[vo.Edges[e0].FFace].ReferencePoint
		averageUV := mgl32.Vec2{float32(p.X / rangeX), float32(p.Y / rangeY)}

		for e1 != sc.EmptyEdge && e1 != e0 {

			v1 = vo.Vertices[vo.Edges[e1].VOrigin].Pos
			v2 := v.Vector{}
			if vo.Edges[vo.Edges[e1].ETwin].VOrigin == sc.EmptyVertex {
				v2 = v.Add(vo.Edges[e1].TmpEdge.Pos, v.Mult(vo.Edges[e1].TmpEdge.Dir, -10.0))
			} else {
				v2 = vo.Vertices[vo.Edges[vo.Edges[e1].ETwin].VOrigin].Pos
			}

			// The assigned averageUV is not correct and must be overwritten later! (Just placeholder now for the real one later)
			mesh = append(mesh, geo.Mesh{mgl32.Vec3{float32(v0.X), float32(v0.Y), float32(v0.Z)}, mgl32.Vec3{0.0, 0.0, 1.0}, averageUV})
			mesh = append(mesh, geo.Mesh{mgl32.Vec3{float32(v1.X), float32(v1.Y), float32(v1.Z)}, mgl32.Vec3{0.0, 0.0, 1.0}, averageUV})
			mesh = append(mesh, geo.Mesh{mgl32.Vec3{float32(v2.X), float32(v2.Y), float32(v2.Z)}, mgl32.Vec3{0.0, 0.0, 1.0}, averageUV})

			e1 = vo.Edges[e1].ENext
		}
	}

	return geo.GenerateGeometryArrayAttributes(&mesh, len(mesh))
}

func createDelaunayEdgesGLBuffer(d sc.Delaunay, rangeX, rangeY float64) geo.ArrayGeometry {
	mesh := make([]geo.Mesh, 0)

	normal := mgl32.Vec3{0.0, 0.0, 1.0}

	dEdges := d.ExtractEdgeList()
	for _, e := range dEdges {
		uv1 := mgl32.Vec2{float32(e.V1.X / rangeX), float32(e.V1.Y / rangeY)}
		uv2 := mgl32.Vec2{float32(e.V1.X / rangeX), float32(e.V1.Y / rangeY)}
		mesh = append(mesh, geo.Mesh{mgl32.Vec3{float32(e.V1.X), float32(e.V1.Y), float32(e.V1.Z)}, normal, uv1})
		mesh = append(mesh, geo.Mesh{mgl32.Vec3{float32(e.V2.X), float32(e.V2.Y), float32(e.V2.Z)}, normal, uv2})
	}

	return geo.GenerateGeometryArrayAttributes(&mesh, len(mesh))
}

func createDelaunayPointsGLBuffer(d sc.Delaunay, rangeX, rangeY float64) geo.Geometry {
	mesh := make([]geo.Mesh, 0)
	indices := make([]uint32, 0)

	normal := mgl32.Vec3{0.0, 0.0, 1.0}

	for i, v := range d.Vertices {
		uv1 := mgl32.Vec2{float32(v.Pos.X / rangeX), float32(v.Pos.Y / rangeY)}
		mesh = append(mesh, geo.Mesh{mgl32.Vec3{float32(v.Pos.X), float32(v.Pos.Y), float32(v.Pos.Z)}, normal, uv1})

		indices = append(indices, uint32(i))
	}

	return geo.GenerateGeometryAttributes(&mesh, &indices, len(mesh), len(indices))
}

func createConvexHullGLBuffer(d sc.Delaunay, rangeX, rangeY float64) geo.Geometry {
	mesh := make([]geo.Mesh, 0)
	indices := make([]uint32, 0)

	normal := mgl32.Vec3{0.0, 0.0, 1.0}

	ch := d.ExtractConvexHull()

	for i, v := range ch {
		uv1 := mgl32.Vec2{float32(v.X / rangeX), float32(v.Y / rangeY)}
		mesh = append(mesh, geo.Mesh{mgl32.Vec3{float32(v.X), float32(v.Y), float32(v.Z)}, normal, uv1})

		indices = append(indices, uint32(i))
	}
	// One extra index to the first element to close the loop!
	indices = append(indices, 0)

	return geo.GenerateGeometryAttributes(&mesh, &indices, len(mesh), len(indices))
}

func createVoronoiEdgesGLBuffer(v sc.Voronoi, rangeX, rangeY float64) geo.ArrayGeometry {
	return createDelaunayEdgesGLBuffer(sc.Delaunay(v), rangeX, rangeY)
}

func createInterpolationControlBuffer(geometry geo.ArrayGeometry) uint {
	var positionBuffer uint
	//glGenBuffers(1, &positionBuffer)
	//glBindBuffer(GL_ARRAY_BUFFER, positionBuffer)
	//glBufferData(GL_ARRAY_BUFFER, PARTICLE_COUNT*sizeof(Vec3), G_ComputePositions, GL_DYNAMIC_COPY)

	return positionBuffer
}

func freeGLBuffers() {
	gl.DeleteBuffers(1, &g_delaunayTriangleGLBuffer.ArrayBuffer)
	gl.DeleteVertexArrays(1, &g_delaunayTriangleGLBuffer.VertexBuffer)

	gl.DeleteBuffers(1, &g_delaunayEdgesGLBuffer.ArrayBuffer)
	gl.DeleteVertexArrays(1, &g_delaunayEdgesGLBuffer.VertexBuffer)

	gl.DeleteBuffers(1, &g_delaunayPointsGLBuffer.ArrayBuffer)
	gl.DeleteBuffers(1, &g_delaunayPointsGLBuffer.IndexBuffer)
	gl.DeleteVertexArrays(1, &g_delaunayPointsGLBuffer.VertexBuffer)

	gl.DeleteBuffers(1, &g_convexHullGLBuffer.ArrayBuffer)
	gl.DeleteBuffers(1, &g_convexHullGLBuffer.IndexBuffer)
	gl.DeleteVertexArrays(1, &g_convexHullGLBuffer.VertexBuffer)

	gl.DeleteBuffers(1, &g_voronoiEdgesGLBuffer.ArrayBuffer)
	gl.DeleteVertexArrays(1, &g_voronoiEdgesGLBuffer.VertexBuffer)
}

/*func redefineProjectionMatrices() {
    gl.UseProgram(g_delaunayTrianglesShader)
    defineMatrices(g_delaunayTrianglesShader)
    defineModelMatrix(g_delaunayTrianglesShader, mgl32.Vec3{-float32(g_windowWidth) / 2, -float32(g_windowHeight) / 2, 0}, mgl32.Vec3{1, 1, 1})

    gl.UseProgram(g_delaunayEdgesShader)
    defineMatrices(g_delaunayEdgesShader)
    defineModelMatrix(g_delaunayEdgesShader, mgl32.Vec3{-float32(g_windowWidth) / 2, -float32(g_windowHeight) / 2, 0}, mgl32.Vec3{1, 1, 1})

    gl.UseProgram(g_delaunayPointsShader)
    defineMatrices(g_delaunayPointsShader)
    defineModelMatrix(g_delaunayPointsShader, mgl32.Vec3{-float32(g_windowWidth) / 2, -float32(g_windowHeight) / 2, 0}, mgl32.Vec3{1, 1, 1})

    gl.UseProgram(0)
}*/

func recalculateDelaunayTriangulation() {

	//g_windowWidth = float64(g_windowWidth)
	//g_windowHeight = float64(g_windowHeight)

	d := createDelaunay(g_delaunayPointCount, float64(g_windowWidth), float64(g_windowHeight), g_delaunayMargin)

	//drawImage(d, "delaunay")

	v := d.CreateVoronoi()

	//drawImage(sc.Delaunay(v), "voronoi")

	//fmt.Println(sc.Delaunay(v))

	freeGLBuffers()

	g_voronoiEdgesGLBuffer = createVoronoiEdgesGLBuffer(v, float64(g_windowWidth), float64(g_windowHeight))

	//g_interpolationControlBuffer = createInterpolationControlBuffer(g_voronoiEdgesGLBuffer)

	g_voronoiTriangleGLBuffer = createVoronoiGLBuffer(v, float64(g_windowWidth), float64(g_windowHeight))
	g_delaunayTriangleGLBuffer = createDelaunayGLBuffer(d, float64(g_windowWidth), float64(g_windowHeight))
	g_delaunayEdgesGLBuffer = createDelaunayEdgesGLBuffer(d, float64(g_windowWidth), float64(g_windowHeight))
	g_delaunayPointsGLBuffer = createDelaunayPointsGLBuffer(d, float64(g_windowWidth), float64(g_windowHeight))
	g_convexHullGLBuffer = createConvexHullGLBuffer(d, float64(g_windowWidth), float64(g_windowHeight))

	//redefineProjectionMatrices()

	gl.MemoryBarrier(gl.ALL_BARRIER_BITS)

}

func defineModelMatrix(shader uint32, pos, scale mgl32.Vec3) {
	matScale := mgl32.Scale3D(scale.X(), scale.Y(), scale.Z())
	matTrans := mgl32.Translate3D(pos.X(), pos.Y(), pos.Z())
	model := matTrans.Mul4(matScale)
	modelUniform := gl.GetUniformLocation(shader, gl.Str("modelMat\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

}

// Defines the Model-View-Projection matrices for the shader.
func defineMatrices(shader uint32) {
	//projection := mgl32.Perspective(g_fovy, g_aspect, g_nearPlane, g_farPlane)
	projection := mgl32.Ortho(-float32(g_windowWidth)/2, float32(g_windowWidth)/2, -float32(g_windowHeight)/2, float32(g_windowHeight)/2, g_nearPlane, g_farPlane)
	camera := mgl32.LookAtV(mgl32.Vec3{0, 0, 1}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})

	viewProjection := projection.Mul4(camera)
	cameraUniform := gl.GetUniformLocation(shader, gl.Str("viewProjectionMat\x00"))
	gl.UniformMatrix4fv(cameraUniform, 1, false, &viewProjection[0])
}

func renderDelaunay() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, g_sceneFboMS)
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.Viewport(0, 0, int32(g_windowWidth), int32(g_windowHeight))

	gl.Disable(gl.DEPTH_TEST)

	//expectedRadius := float32(g_windowWidth / g_windowWidth / math.Sqrt(float64(g_delaunayPointCount)))
	expectedRadius := float32(calcExpectedRadius(g_delaunayPointCount, float64(g_windowWidth), float64(g_windowHeight), g_delaunayMargin))
	expectedRadiusX := expectedRadius / float32(g_windowWidth)
	expectedRadiusY := expectedRadius / float32(g_windowHeight)

	if g_renderTriangles {
		gl.UseProgram(g_delaunayTrianglesShader)
		gl.BindVertexArray(g_delaunayTriangleGLBuffer.VertexBuffer)
		gl.Uniform1f(gl.GetUniformLocation(g_delaunayTrianglesShader, gl.Str("expectedRadiusX\x00")), expectedRadiusX)
		gl.Uniform1f(gl.GetUniformLocation(g_delaunayTrianglesShader, gl.Str("expectedRadiusY\x00")), expectedRadiusY)
		gl.DrawArrays(gl.TRIANGLES, 0, g_delaunayTriangleGLBuffer.VertexCount)
	}

	if g_renderVoronoiCells {
		gl.UseProgram(g_delaunayTrianglesShader)
		gl.BindVertexArray(g_voronoiTriangleGLBuffer.VertexBuffer)
		gl.Uniform1f(gl.GetUniformLocation(g_delaunayTrianglesShader, gl.Str("expectedRadiusX\x00")), expectedRadiusX)
		gl.Uniform1f(gl.GetUniformLocation(g_delaunayTrianglesShader, gl.Str("expectedRadiusY\x00")), expectedRadiusY)
		gl.DrawArrays(gl.TRIANGLES, 0, g_voronoiTriangleGLBuffer.VertexCount)

	}

	if g_renderLines {
		gl.UseProgram(g_delaunayEdgesShader)
		gl.BindVertexArray(g_delaunayEdgesGLBuffer.VertexBuffer)
		gl.Uniform1i(gl.GetUniformLocation(g_delaunayEdgesShader, gl.Str("useExternalColor\x00")), g_useExternalColor)
		gl.Uniform3fv(gl.GetUniformLocation(g_delaunayEdgesShader, gl.Str("color\x00")), 1, &g_delaunayLineColor[0])
		gl.DrawArrays(gl.LINES, 0, g_delaunayEdgesGLBuffer.VertexCount)
	}

	if g_renderPoints {
		gl.UseProgram(g_delaunayPointsShader)
		gl.BindVertexArray(g_delaunayPointsGLBuffer.VertexBuffer)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, g_delaunayPointsGLBuffer.IndexBuffer)
		gl.Uniform1i(gl.GetUniformLocation(g_delaunayPointsShader, gl.Str("useExternalColor\x00")), g_useExternalColor)
		gl.Uniform3fv(gl.GetUniformLocation(g_delaunayPointsShader, gl.Str("color\x00")), 1, &g_pointColor[0])
		gl.DrawElements(gl.POINTS, g_delaunayPointsGLBuffer.IndexCount, gl.UNSIGNED_INT, gl.PtrOffset(0))
	}

	if g_renderConvexHull {
		gl.UseProgram(g_delaunayEdgesShader)
		gl.BindVertexArray(g_convexHullGLBuffer.VertexBuffer)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, g_convexHullGLBuffer.IndexBuffer)
		gl.Uniform1i(gl.GetUniformLocation(g_delaunayEdgesShader, gl.Str("useExternalColor\x00")), 1)
		gl.Uniform3fv(gl.GetUniformLocation(g_delaunayEdgesShader, gl.Str("color\x00")), 1, &g_chColor[0])
		gl.DrawElements(gl.LINE_STRIP, g_convexHullGLBuffer.IndexCount, gl.UNSIGNED_INT, gl.PtrOffset(0))
	}

	if g_renderVoronoiEdges {
		gl.UseProgram(g_delaunayEdgesShader)
		gl.BindVertexArray(g_voronoiEdgesGLBuffer.VertexBuffer)
		gl.Uniform1i(gl.GetUniformLocation(g_delaunayEdgesShader, gl.Str("useExternalColor\x00")), 1)
		gl.Uniform3fv(gl.GetUniformLocation(g_delaunayEdgesShader, gl.Str("color\x00")), 1, &g_voronoiLineColor[0])
		gl.DrawArrays(gl.LINES, 0, g_voronoiEdgesGLBuffer.VertexCount)
	}

	// Multisampling
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, g_sceneFboMS)
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
	gl.DrawBuffer(gl.BACK)
	gl.BlitFramebuffer(0, 0, int32(g_windowWidth), int32(g_windowHeight), 0, 0, int32(g_windowWidth), int32(g_windowHeight), gl.COLOR_BUFFER_BIT|gl.DEPTH_BUFFER_BIT, gl.NEAREST)

}

func prepareGLForNewTexture(imagePath string) {

	// OpenGL should silently ignore if the texture doesn't exist!
	gl.DeleteTextures(1, &g_delaunayTexture.TextureHandle)
	gl.DeleteTextures(1, &g_sceneColorTexMS)
	gl.DeleteTextures(1, &g_sceneDepthTexMS)
	gl.DeleteFramebuffers(1, &g_sceneFboMS)

	g_delaunayTexture = mtgl.CreateImageTexture(imagePath, false)

	g_windowWidth = int(g_delaunayTexture.TextureSize.X())
	g_windowHeight = int(g_delaunayTexture.TextureSize.Y())

	// Make sure the window will not get really large just because of the loaded texture.
	if g_windowWidth >= g_windowHeight && g_windowWidth > g_maxWindowSize {
		g_windowHeight = int(float64(g_windowHeight) * g_maxWindowSize / float64(g_windowWidth))
		g_windowWidth = g_maxWindowSize
	}
	if g_windowHeight > g_windowWidth && g_windowHeight > g_maxWindowSize {
		g_windowWidth = int(float64(g_windowWidth) * g_maxWindowSize / float64(g_windowHeight))
		g_windowHeight = g_maxWindowSize
	}

	g_window.SetSize(g_windowWidth, g_windowHeight)

	g_sceneFboMS = mtgl.CreateFbo(&g_sceneColorTexMS, &g_sceneDepthTexMS, int32(g_windowWidth), int32(g_windowHeight), true, 16, false, 1)

	gl.UseProgram(g_delaunayTrianglesShader)
	defineMatrices(g_delaunayTrianglesShader)
	defineModelMatrix(g_delaunayTrianglesShader, mgl32.Vec3{-float32(g_windowWidth) / 2, -float32(g_windowHeight) / 2, 0}, mgl32.Vec3{1, 1, 1})
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, g_delaunayTexture.TextureHandle)
	gl.Uniform1i(gl.GetUniformLocation(g_delaunayTrianglesShader, gl.Str("imageTexture\x00")), 0)

	gl.UseProgram(g_delaunayEdgesShader)
	defineMatrices(g_delaunayEdgesShader)
	defineModelMatrix(g_delaunayEdgesShader, mgl32.Vec3{-float32(g_windowWidth) / 2, -float32(g_windowHeight) / 2, 0}, mgl32.Vec3{1, 1, 1})
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, g_delaunayTexture.TextureHandle)
	gl.Uniform1i(gl.GetUniformLocation(g_delaunayEdgesShader, gl.Str("imageTexture\x00")), 0)

	gl.UseProgram(g_delaunayPointsShader)
	defineMatrices(g_delaunayPointsShader)
	defineModelMatrix(g_delaunayPointsShader, mgl32.Vec3{-float32(g_windowWidth) / 2, -float32(g_windowHeight) / 2, 0}, mgl32.Vec3{1, 1, 1})
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, g_delaunayTexture.TextureHandle)
	gl.Uniform1i(gl.GetUniformLocation(g_delaunayPointsShader, gl.Str("imageTexture\x00")), 0)

	gl.UseProgram(0)
}

func SetPointDistributionMethod(method int) {
	g_delaunayDistribution = method
}

func IncreasePointCount() {
	g_delaunayPointCount *= 2
}

func DecreasePointCount() {
	g_delaunayPointCount /= 2
	if g_delaunayPointCount < 3 {
		g_delaunayPointCount = 3
	}
}
func SetRenderVoronoiCells(show bool) {
	g_renderVoronoiCells = show
}
func SetRenderTriangles(show bool) {
	g_renderTriangles = show
}
func SetRenderVoronoiEdges(show bool) {
	g_renderVoronoiEdges = show
}
func SetRenderLines(show bool) {
	g_renderLines = show
}
func SetRenderPoints(show bool) {
	g_renderPoints = show
}
func SetRenderConvexHull(show bool) {
	g_renderConvexHull = show
}
func SetUseExternalColor(useExternalColor bool) {
	if useExternalColor {
		g_useExternalColor = 1
	} else {
		g_useExternalColor = 0
	}
}
func SetNewImage(path string) {
	prepareGLForNewTexture(path)
}
func SaveImage(path string) {

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	pixels := make([]byte, g_windowWidth*g_windowHeight*4)
	pixelsFlipped := make([]byte, g_windowWidth*g_windowHeight*4)
	gl.ReadBuffer(gl.FRONT)
	gl.ReadPixels(0, 0, int32(g_windowWidth), int32(g_windowHeight), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pixels))

	// Flipping the Y-Axis because OpenGL's y-axis is mirrored compared to normal images.
	// There might be a more efficient way to do this, but for now this works fine and is not too slow (doesn't happen very often anyway)
	for x := 0; x < g_windowWidth; x++ {
		for y := 0; y < g_windowHeight; y++ {
			pixelsFlipped[(y*g_windowWidth+x)*4] = pixels[((g_windowHeight-(y+1))*g_windowWidth+x)*4]
			pixelsFlipped[(y*g_windowWidth+x)*4+1] = pixels[((g_windowHeight-(y+1))*g_windowWidth+x)*4+1]
			pixelsFlipped[(y*g_windowWidth+x)*4+2] = pixels[((g_windowHeight-(y+1))*g_windowWidth+x)*4+2]
			pixelsFlipped[(y*g_windowWidth+x)*4+3] = pixels[((g_windowHeight-(y+1))*g_windowWidth+x)*4+3]
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, g_windowWidth, g_windowHeight))
	img.Pix = pixelsFlipped

	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		fmt.Printf("error when creating an image: %v\n", err)
	}
	err = png.Encode(file, img)
	if err != nil {
		fmt.Printf("error when saving an image: %v\n", err)
	}

}
func SetVoronoiLineColor(r, g, b, a float64) {
	g_voronoiLineColor = mgl32.Vec4{float32(r), float32(g), float32(b), float32(a)}
}
func SetDelaunayLineColor(r, g, b, a float64) {
	g_delaunayLineColor = mgl32.Vec4{float32(r), float32(g), float32(b), float32(a)}
}
func SetPointColor(r, g, b, a float64) {
	g_pointColor = mgl32.Vec4{float32(r), float32(g), float32(b), float32(a)}
}
func SetCHColor(r, g, b, a float64) {
	g_chColor = mgl32.Vec4{float32(r), float32(g), float32(b), float32(a)}
}

func CloseWindow() {
	g_window.SetShouldClose(true)
}

func ReadyForRebuild(r bool) {
	g_readyForRebuild = r
}
func ReadyForRender(r bool) {
	g_readyForRender = r
}

// Register all needed callbacks
func registerCallBacks(window *glfw.Window) {

	window.SetKeyCallback(func(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press {
			switch key {
			// Close the Simulation.
			case glfw.KeyEscape, glfw.KeyQ:
				window.SetShouldClose(true)
			case glfw.KeyF2:
			case glfw.KeyLeft:
			case glfw.KeyRight:
			}

		}
	})
}

func pollCommunicationChannel() {

	// Timeout of the blocking channel read about 10 times per second to allow glfw window updates to happen without using heaps of CPU.
	select {
	case msg := <-g_controlCommunication:
		msg()
	case <-time.After(100 * time.Millisecond):
	}
}

// Mainloop for communication, fps and keyboard input.
// Re-render will be triggered by the controls!
func mainLoop(window *glfw.Window) {

	registerCallBacks(window)
	glfw.SwapInterval(1)

	gl.PointSize(5)
	gl.Enable(gl.LINE_SMOOTH)

	for !window.ShouldClose() {

		g_currentTime = glfw.GetTime()

		if g_readyForRebuild {
			recalculateDelaunayTriangulation()
			g_readyForRebuild = false
		}

		if g_readyForRender {
			renderDelaunay()
			g_readyForRender = false
			window.SwapBuffers()
		}

		glfw.PollEvents()

		pollCommunicationChannel()
	}

	// We are closing the GLFW window. Signaling the closing to our control window so it can kill itself.
	select {
	case g_controlCommunicationClose <- 1:
	default:
	}
}

// Set OpenGL version, profile and compatibility
func initGraphicContext() (*glfw.Window, error) {
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	//glfw.WindowHint(glfw.Decorated, glfw.False)

	window, err := glfw.CreateWindow(g_windowWidth, g_windowHeight, g_windowTitle, nil, nil)
	if err != nil {
		return nil, err
	}
	window.MakeContextCurrent()

	// Initialize Glow
	if err := gl.Init(); err != nil {
		return nil, err
	}

	return window, nil
}

func InitializeRender(communication chan func(), closingChannel chan int, pointCount int, defaultImage string) {
	var err error = nil
	if err = glfw.Init(); err != nil {
		panic(err)
	}
	// Terminate as soon, as this the function is finished.
	defer glfw.Terminate()

	window, err := initGraphicContext()
	if err != nil {
		// Decision to panic or do something different is taken in the main
		// method and not in sub-functions
		panic(err)
	}

	path := "./"
	g_delaunayTrianglesShader, err = mtgl.NewProgram(path+"triangles.vert", "", "", "", path+"simple.frag")
	if err != nil {
		panic(err)
	}
	g_delaunayEdgesShader, err = mtgl.NewProgram(path+"simple.vert", path+"edges.geo", "", "", path+"simple.frag")
	if err != nil {
		panic(err)
	}
	g_delaunayPointsShader, err = mtgl.NewProgram(path+"points.vert", "", "", "", path+"points.frag")
	if err != nil {
		panic(err)
	}

	g_delaunayPointCount = pointCount
	g_controlCommunication = communication
	g_controlCommunicationClose = closingChannel
	g_window = window

	prepareGLForNewTexture(defaultImage)

	mainLoop(window)
}
