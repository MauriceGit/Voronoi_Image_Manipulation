## Voronoi/Delaunay image manipulation

This project implements a small, lightweight tool for real-time image manipulation with Voronoi/Delaunay data structures.

The focus of this project was, to enable users without knowledge of voronoi or delaunay specifics, to experiment and play with different looks a voronoi or delaunay structure can give to an image.

Some time ago I implemented a similar effect (also voronoi/delaunay) in Python ([github.com/MauriceGit/Delaunay_Triangulation](https://github.com/MauriceGit/Delaunay_Triangulation)). The main goal for this project was, to make it a lot more robust, user friendly, fast and usable.

## Interface:

The program will start two separate windows. One to actually display the image and a control window. It will look like the following:

Image view                 |  Control view
:-------------------------:|:-------------------------:
![Image view](Screenshots/view_default.png)  |  ![Controls](Screenshots/controls.png)

## Requirements:

- Graphics card supporting OpenGL 3.3
- Windows users: Have mingw64 installed. Other C compilers might or might not work. Please report back if it works or submit necessary changes to this README.

## Run:

- run ```go get ./...``` to install all dependencies
- run ```go build``` within the projects directory
- Do not remove the _Images/apple.png_ directory. This image is loaded by default when the program starts.

## Screenshots and usecases:

Just to give you and incomplete overview what kind of effects you can achieve with this tool (sometimes with the corresponding controls set).

The _point Distribution_ set to _Poisson Disk_ to achieve random but equally distributed points over the whole area. This gives the most pleasing and homogeneous look most of the time.
![Poisson Disk point dist](Screenshots/apple_poisson.png)

The _point Distribution_ set to _Random_. Truly random point distribution. Will create unequally sized regions.
![Random point dist](Screenshots/apple_random.png)

The _point Distribution_ set to _Grid_ will create honeycomb like regions (hexagons) by placing points in a shifted grid.
![Grid point dist](Screenshots/apple_grid.png)

![Grid point dist](Screenshots/voronoi_grid.png)

Set the _Face Rendering_ to _Delaunay Triangles_.
![Delaunay faces](Screenshots/delaunay.png)

![Delaunay faces](Screenshots/delaunay_tiger.png)

An image of a Labrador with Poisson disk distributed points.
![Voronoi faces](Screenshots/voronoi_controls.png)

If you like you can add the voronoi lines and points to actually display the underlaying data structure.
![Voronoi lines and points](Screenshots/voronoi_lines_points.png)

![Voronoi lines](Screenshots/ara_voronoi.png)

![Voronoi lines](Screenshots/voronoi_rose.png)

Sunset with Voronoi        |  Sunset with Delaunay
:-------------------------:|:-------------------------:
![Sunset voronoi](Screenshots/voronoi_sunset.png)  |  ![Sunset delaunay](Screenshots/delaunay_sunset.png)

When checking _Adaptive Color_, the lines of delaunay edges and points will get the average color of the image underneath. Creating an interesting effect.
![Adaptive Color](Screenshots/adaptive_ara.png)

_Adaptive Color_ checked with Voronoi faces and points displayed with a grid layout.
![Adaptive Color voronoi](Screenshots/adaptive_points_voronoi.png)

You could just view the points as well (well that doesn't really use voronoi/delaunay any more. But still looks cool :)).
![Just points](Screenshots/points_ara.png)

At last, you can just ignore the image to investigate/look at voronoi and delaunay tessellations itself:
![Just voronoi](Screenshots/random_voronoi.png)

![Just voronoi and delaunay](Screenshots/voronoi_delaunay_controls.png)
