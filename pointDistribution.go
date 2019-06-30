// pointDistribution
package main

import (
	//"fmt"
	"math"
	"math/rand"

	//v "github.com/MauriceGit/mtVector"
	sc "mtSweepCircle"
)

func calcExpectedRadius(count int, rangeX, rangeY, margin float64) float64 {
	ratio := (rangeX - 2*margin) / (rangeY - 2*margin)
	square := (rangeX - 2*margin) / ratio
	// How many points should fit in the square
	newCount := float64(count) / ratio
	return square / math.Sqrt(newCount)
}

func CreateGridPoints(count int, rangeX, rangeY, margin float64) []sc.Vector {
	var pointList []sc.Vector

	r := calcExpectedRadius(count, rangeX, rangeY, margin)

	for i := margin; i <= rangeX-margin; i += r {
		for j := margin; j <= rangeY-margin; j += r {
			pointList = append(pointList, sc.Vector{i, j})
		}
	}

	return pointList
}

func CreateShiftedGridPoints(count int, rangeX, rangeY, margin float64) []sc.Vector {
	var pointList []sc.Vector

	r := calcExpectedRadius(count, rangeX, rangeY, margin)

	shift := 0.0
	loopCount := 0

	for i := margin; i <= rangeX-margin; i += r {

		if loopCount%2 == 0 {
			shift = 0.0
		} else {
			shift = r / 2.0
		}

		for j := margin; j <= rangeY-margin; j += r {
			if j+shift <= rangeY-margin {
				pointList = append(pointList, sc.Vector{i, j + shift})
			}
		}

		loopCount++
	}

	return pointList
}

func CreateRandomPoints(count int, rangeX, rangeY, margin float64, seed int64) []sc.Vector {
	//var seed int64 = time.Now().UTC().UnixNano()
	r := rand.New(rand.NewSource(seed))
	var pointList []sc.Vector

	for i := 0; i < count; i++ {
		v := sc.Vector{r.Float64()*(rangeX-2*margin) + margin, r.Float64()*(rangeY-2*margin) + margin}
		pointList = append(pointList, v)
	}
	return pointList
}

func randVec(base sc.Vector, minR, maxR float64, rd *rand.Rand) sc.Vector {
	vx := rd.Float64()*(maxR-minR) + minR
	vy := 0.0
	a := rd.Float64() * 2.0 * math.Pi
	return sc.Vector{base.X + vx*math.Cos(a) - vy*math.Sin(a), base.Y + vy*math.Cos(a) + vx*math.Sin(a)}
}

func getGridPos(p sc.Vector, f float64) (int, int) {
	return int(math.Floor(p.X / f)), int(math.Floor(p.Y / f))
}

func fits(grid *[]int, pointList []sc.Vector, p sc.Vector, r float64, gx, gy, width, height int) bool {
	for i := int(math.Max(float64(gx)-2, 0)); i < int(math.Min(float64(gx)+3, float64(width))); i++ {
		for j := int(math.Max(float64(gy)-2, 0)); j < int(math.Min(float64(gy)+3, float64(height))); j++ {
			pg := (*grid)[i+j*width]
			if pg != -1 {
				pgp := pointList[pg]
				if sc.Length(sc.Sub(p, pgp)) <= r {
					return false
				}
			}
		}
	}
	return true
}

func CreateFastPoissonDiscPoints(count int, rangeX, rangeY, margin float64, k int, seed int64) []sc.Vector {

	r := calcExpectedRadius(count, rangeX, rangeY, margin)

	//var seed int64 = time.Now().UTC().UnixNano()
	rd := rand.New(rand.NewSource(seed))

	cellSize := r / math.Sqrt(2)
	gridWidth := int(math.Ceil(rangeX / cellSize))
	gridHeight := int(math.Ceil(rangeY / cellSize))

	grid := make([]int, gridWidth*gridHeight)
	for i := 0; i < gridWidth*gridHeight; i++ {
		grid[i] = -1
	}

	var pointList []sc.Vector
	var activeList []sc.Vector

	p := sc.Vector{rd.Float64()*(rangeX-2*margin) + margin, rd.Float64()*(rangeY-2*margin) + margin}

	pointList = append(pointList, p)
	activeList = append(activeList, p)
	gridX, gridY := getGridPos(p, cellSize)
	grid[gridX+gridY*gridWidth] = 0

	for len(activeList) > 0 && len(pointList) < count {

		qi := rd.Intn(len(activeList))
		q := activeList[qi]
		activeList[qi] = activeList[len(activeList)-1]
		activeList = activeList[:len(activeList)-1]

		for tmp := 0; tmp < k && len(pointList) < count; tmp++ {
			p = randVec(q, r, 2.0*r, rd)

			if p.X >= margin && p.X < rangeX-margin && p.Y >= margin && p.Y < rangeY-margin {
				gridX, gridY = getGridPos(p, cellSize)
				if fits(&grid, pointList, p, r, gridX, gridY, gridWidth, gridHeight) {
					activeList = append(activeList, p)
					pointList = append(pointList, p)
					grid[gridX+gridY*gridWidth] = len(pointList) - 1
				}
			}
		}
	}
	return pointList

}
