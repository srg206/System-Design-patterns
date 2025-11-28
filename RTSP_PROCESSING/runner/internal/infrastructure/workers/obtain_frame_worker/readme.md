


1) start:  mediamtx mediamtx.yml 
2) run any videofile 
ffmpeg -re -stream_loop -1 -i "video.mp4" \
  -c copy \
  -f rtsp \
  -rtsp_transport tcp \
  rtsp://localhost:8554/mystream

3) Now you can run test 
NOTE: these are not normal integration or e2e test
      Its just for local development to make sure workers pipline is working

4) To test S3UploadDownload create in this folder test_frame.jpg

  cd obtain_frame_worker
  go test -run TestRTSPStreamGrabFrameWorker -v ./...
  go test -run TestRTSPStreamParsing -v ./...
  go test -run TestS3UploadAndDownloadFrame -v ./...