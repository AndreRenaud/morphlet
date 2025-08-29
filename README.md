# Convert PNGs to .mp4

ffmpeg -y -f image2 -framerate 10 -i warped-%05d.png -vcodec libx264  -pix_fmt yuv420p video.mp4