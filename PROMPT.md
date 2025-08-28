This is a program that takes an ordered series of images, and then lets you compare each image with the ones before/after it. In that comparison, similar points can be added/removed, and then once all these comparison points are present it will construct a movie morphing the images.

The UI should be done using github.com/AllenDang/giu
The image morphing should be done with gocv.io/x/gocv

The UI should start with a simple screen saying "Start New Project" and "Open Project"

Once a project is created, there should be a list on the left hand side where images can be added/removed and reordered.
When an image is selected, it is displayed on the right hand side, along with the image before it in the list (if you select the first one, it should just say "please select the next image")

When the images are displayed beside each other, points can be added by clicking on either image (a corresponding point will be placed on the other image), then dragged around. Right clicking will delete the point, and it's pair on the other image.