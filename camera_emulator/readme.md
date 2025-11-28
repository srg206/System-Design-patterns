1) mediamtx mediamtx.yml 
2) run any videofile:

ffmpeg -re -stream_loop -1 -i "video.mp4" \
  -c copy \
  -f rtsp \
  -rtsp_transport tcp \
  rtsp://localhost:8554/mystream