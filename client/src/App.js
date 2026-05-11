import React, { useState, useEffect } from "react";

const App = () => {
  const [videos, setVideos] = useState([]);
  const [selectedVideo, setSelectedVideo] = useState(null);

  useEffect(() => {
    fetch("http://localhost:8080/videos")
      .then((res) => {
        if (!res.ok) throw new Error("Failed to fetch video list");
        return res.json();
      })
      .then((data) => {
        console.log("Backend ответил:", data);
        setVideos(Array.isArray(data.videos) ? data.videos : []);
      })
      .catch((err) => {
        console.error("Error fetching videos:", err);
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
      .then((res) => res.text())
      .then((msg) => {
        alert(msg);
        return fetch("http://localhost:8080/videos");
      })
      .then((res) => res.json())
      .then((data) => setVideos(Array.isArray(data.videos) ? data.videos : []))
      .catch((err) => alert("Error: " + err.message));
  };

  return (
    <div style={{ padding: "20px", fontFamily: "sans-serif" }}>
      <h1>Video Streaming Service</h1>
      
      <h2>Available Videos</h2>
      <ul>
        {videos.length > 0 ? (
          videos.map((video) => (
            <li key={video} style={{ marginBottom: "8px" }}>
              <button onClick={() => setSelectedVideo(video)}>{video}</button>
            </li>
          ))
        ) : (
          <p>No videos available</p>
        )}
      </ul>

      {selectedVideo && (
        <div style={{ marginTop: "20px" }}>
          <h2>Now Playing: {selectedVideo}</h2>
          <video controls autoPlay style={{ width: "100%", maxWidth: "800px", background: "#000" }}>
            <source src={`http://localhost:8080/stream?file=${selectedVideo}`} type="video/mp4" />
            Your browser does not support the video tag.
          </video>
        </div>
      )}

      <h2 style={{ marginTop: "30px" }}>Upload Video</h2>
      <form onSubmit={uploadVideo}>
        <input type="file" name="video" accept="video/mp4,video/mkv" />
        <button type="submit" style={{ marginLeft: "10px" }}>Upload</button>
      </form>
    </div>
  );
};

export default App;