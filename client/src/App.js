import React, { useState, useEffect } from "react";

const App = () => {
  const [videos, setVideos] = useState([]);
  const [selectedVideo, setSelectedVideo] = useState(null);

  useEffect(() => {
    fetch("http://localhost:8080/videos")
      .then((response) => {
        if (!response.ok) {
          throw new Error("Failed to fetch video list");
        }
        return response.json();
      })
      .then((data) => {
        console.log("Fetched videos:", data.videos);
        setVideos(Array.isArray(data.videos) ? data : []);
      })
      .catch((error) => {
        console.error("Error fetching videos:", error.message);
        alert(error.message);
        setVideos([]);
      });
  }, []);

  const uploadVideo = (event) => {
    event.preventDefault();
    const file = event.target.video.files[0];

    if (!file) {
      alert("Please select a video file to upload.");
      return;
    }

    const formData = new FormData();
    formData.append("video", file);

    fetch("http://localhost:8080/upload", { method: "POST", body: formData })
      .then((response) => response.text())
      .then((data) => {
        alert(data);
        return fetch("http://localhost:8080/videos");
      })
      .then((response) => {
        if (!response.ok) {
          throw new Error("Failed to fetch updated video list");
        }
        return response.json();
      })
      .then((updatedVideos) => setVideos(Array.isArray(updatedVideos) ? updatedVideos : []))
      .catch((error) =>
        alert("Error uploading file or fetching updated list: " + error.message)
      );
  };

  return (
    <div>
      <h1>Video Streaming Service</h1>
      <h2>Available Videos</h2>
      <ul>
        {(Array.isArray(videos.videos) ? videos.videos : []).map((video) => (
          <li key={video}>
            <button onClick={() => setSelectedVideo(video)}>{video}</button>
          </li>
        ))}
      </ul>
      {selectedVideo && (
        <video controls autoPlay>
          <source
            src={`http://localhost:8080/stream?file=${selectedVideo}`}
            type="video/mp4"
          />
        </video>
      )}
      <h2>Upload Video</h2>
      <form onSubmit={uploadVideo}>
        <input type="file" name="video" accept="video/mp4,video/mkv" />
        <button type="submit">Upload</button>
      </form>
    </div>
  );
};

export default App;
