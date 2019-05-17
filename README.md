## Voronoi/Delaunay image manipulation

This project implements a small, lightweight tool for real-time image manipulation with Voronoi/Delaunay data structures.

The focus of this project was, to enable users without knowledge of voronoi or delaunay specifics, to experiment and play with different looks a voronoi or delaunay structure can give an image.

Some time ago I implemented a similar effect (also voronoi/delaunay) in Python ([github.com/MauriceGit/Delaunay_Triangulation](https://github.com/MauriceGit/Delaunay_Triangulation)). The main goal for this project was, to make it a lot more robust, user friendly, fast and usable.

## Interface:

The program will start two separate windows. One to actually display the image and a control window. It will look like the following:

Image view             |  Control view
:-------------------------:|:-------------------------:
![Image view](Screenshots/voronoi_sunset.png)  |  ![Controls](Screenshots/controls.png)

## Requirements:

- Graphics card supporting OpenGL 3.3
- Windows users: Have mingw64 installed. Other C compilers might or might not work. Please report back if it works or submit necessary changes to this README.

## Run:

- run ```go get ./...``` to install all dependencies
- run ```go build``` within the projects directory
- Do not remove the _Images/apple.png_ directory. This image is loaded by default when the program starts.

## Screenshots and usecases:
