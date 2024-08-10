use warp::Filter;
use serde::{Deserialize, Serialize};
use std::process::Command;
use std::time::Instant;
use log::{debug, info, error};

use google_cloud_storage::client::{Client, ClientConfig};
use google_cloud_storage::http::objects::upload::{UploadObjectRequest, UploadType};
use google_cloud_storage::http::objects::Object;
use google_cloud_storage::http::object_access_controls::PredefinedObjectAcl;
use anyhow::{Result, Context, anyhow};

use std::fs;
use std::path::{PathBuf, Path};
use std::env;
use tokio::fs::File;
use tokio::io::AsyncReadExt;
use std::collections::HashMap;

const REEL_ASPECT_RATIO: f32 = 9.0 / 16.0;
const REEL_WIDTH: u32 = 1080;
const REEL_HEIGHT: u32 = (REEL_WIDTH as f32 / REEL_ASPECT_RATIO) as u32;

#[derive(Debug, Deserialize, Serialize)]
struct Word {
    start: f64,
    end: f64,
    word: String,
}

#[derive(Debug, Deserialize, Serialize)]
struct Sentence {
    start: f64,
    end: f64,
    text: String,
}

#[derive(Debug, Deserialize, Serialize)]
struct ASRData {
    sentences: Vec<Sentence>,
    words: Vec<Word>,
}

#[derive(Debug, Deserialize)]
struct CreateSlideshowRequest {
    video_id: String,
}

#[derive(Debug, Serialize)]
struct CreateSlideshowResponse {
    message: String,
    output_file: String,
}

async fn upload_video_to_gcs(video_id: &str, output_file: &str) -> Result<()> {
    let config = ClientConfig::default().with_auth().await.context("Failed to create client config")?;
    let client = Client::new(config);
    let bucket_name = "zappush_public";
    let object_name = format!("videos/{}/full_video.mp4", video_id);

    let mut file = File::open(output_file).await.context("Failed to open file")?;
    let mut buffer = Vec::new();
    file.read_to_end(&mut buffer).await.context("Failed to read file")?;

    let upload_request = UploadObjectRequest {
        bucket: bucket_name.to_string(),
        predefined_acl: Some(PredefinedObjectAcl::PublicRead),
        ..Default::default()
    };

    // Create metadata
    let mut metadata = HashMap::new();
    metadata.insert("video_id".to_string(), video_id.to_string());

    // Create the upload type with metadata
    let upload_type = UploadType::Multipart(Box::new(Object {
        name: object_name.clone(),
        content_type: Some("video/mp4".to_string()),
        metadata: Some(metadata),
        ..Default::default()
    }));
    
    // Perform the upload
    client.upload_object(&upload_request, buffer, &upload_type)
        .await
        .context("Failed to upload object")?;
    
    println!("Uploaded video to GCS: {}", object_name);
    Ok(())
}

fn get_video_folder_path(video_id: &str) -> PathBuf {
    let home_dir = env::var("HOME").expect("HOME environment variable not set");
    PathBuf::from(home_dir).join("Desktop").join("reels").join(video_id)
}

async fn create_slideshow(req: CreateSlideshowRequest) -> Result<impl warp::Reply, warp::Rejection> {
    let result: Result<CreateSlideshowResponse, anyhow::Error> = async {
        println!("Creating slideshow for video ID: {}", req.video_id);

        let video_folder = get_video_folder_path(&req.video_id);
        let subtitles_path = video_folder.join("subtitles/subtitles.json");
        let audio_file = video_folder.join("audio/full_audio.mp3");
        let output_file = video_folder.join("output_rust.mp4");

        // Read and parse the subtitles.json file
        let asr_data: ASRData = serde_json::from_str(&fs::read_to_string(&subtitles_path)
            .context("Failed to read subtitles.json")?).context("Failed to parse subtitles.json")?;

        // Get image paths
        let image_paths: Vec<PathBuf> = fs::read_dir(video_folder.join("images"))
            .context("Failed to read images directory")?
            .filter_map(|entry| {
                let entry = entry.ok()?;
                let path = entry.path();
                if path.is_file() && path.file_name()?.to_str()?.starts_with("image_") {
                    Some(path)
                } else {
                    None
                }
            })
            .collect();

        println!("ASR data: {:?}", asr_data);
        println!("Audio file: {}", audio_file.display());
        println!("Output file: {}", output_file.display());

        create_slideshow_with_subtitles(&image_paths, &asr_data, audio_file.to_str().unwrap(), output_file.to_str().unwrap())
            .context("Failed to create slideshow")?;

        println!("Slideshow created successfully");
        upload_video_to_gcs(&req.video_id, output_file.to_str().unwrap())
            .await
            .context("Failed to upload video to GCS")?;

        // guess the URL of the uploaded video
        let url = format!("https://storage.googleapis.com/zappush_public/videos/{}/full_video.mp4", req.video_id);
        println!("Uploaded video to GCS");
        
        Ok(CreateSlideshowResponse {
            message: "Slideshow created successfully".to_string(),
            output_file: url,
        })
    }.await;

    match result {
        Ok(response) => Ok(warp::reply::json(&response)),
        Err(e) => {
            let error_message = format!("Error: {}", e);
            error!("{}", error_message);
            Ok(warp::reply::json(&CreateSlideshowResponse {
                message: error_message,
                output_file: "".to_string(),
            }))
        }
    }
}

fn create_slideshow_with_subtitles(
    image_paths: &[PathBuf],
    asr_data: &ASRData,
    audio_file: &str,
    output_file: &str,
) -> Result<()> {
    let start_time = Instant::now();

    // Ensure the output directory exists
    if let Some(parent) = Path::new(output_file).parent() {
        fs::create_dir_all(parent).context("Failed to create output directory")?;
        println!("Created output directory: {:?}", parent);
    }

    // Create subtitle file
    create_subtitle_file(asr_data, "/tmp/subtitles.srt").context("Failed to create subtitle file")?;
    println!("Created subtitle file");

    // Sort image paths
    let mut sorted_image_paths = image_paths.to_vec();
    sorted_image_paths.sort_by(|a, b| {
        let a_num = a.file_stem().unwrap().to_str().unwrap().split('_').last().unwrap().parse::<u32>().unwrap();
        let b_num = b.file_stem().unwrap().to_str().unwrap().split('_').last().unwrap().parse::<u32>().unwrap();
        a_num.cmp(&b_num)
    });

    // Prepare FFmpeg command
    let mut ffmpeg_args = vec![
        "-y".to_string(),  // Overwrite output file if it exists
    ];

    // Add input images and audio
    for path in &sorted_image_paths {
        ffmpeg_args.extend(vec![
            "-loop".to_string(),
            "1".to_string(),
            "-i".to_string(),
            path.to_str().unwrap().to_string()
        ]);
    }
    ffmpeg_args.extend(vec!["-i".to_string(), audio_file.to_string()]);

    // Create filter complex
    let mut filter_complex = String::new();
    for i in 0..sorted_image_paths.len() {
        filter_complex.push_str(&format!(
            "[{}:v]scale={}:{}:force_original_aspect_ratio=increase,crop={}:{},setsar=1[v{}];", 
            i, REEL_WIDTH, REEL_HEIGHT, REEL_WIDTH, REEL_HEIGHT, i
        ));
    }

    // Create timeline for images
    let mut timeline = String::new();
    let total_duration = asr_data.sentences.last().unwrap().end;
    for (i, sentence) in asr_data.sentences.iter().enumerate() {
        let start = if i == 0 { 0.0 } else { asr_data.sentences[i-1].end };
        let duration = sentence.end - start;
        timeline.push_str(&format!("[v{}]trim=0:{},setpts=PTS-STARTPTS[v{}trim];", i, duration, i));
    }
    timeline.push_str(&format!("{}concat=n={}:v=1:a=0[outv];", 
        (0..sorted_image_paths.len()).map(|i| format!("[v{}trim]", i)).collect::<Vec<_>>().join(""), 
        sorted_image_paths.len()));

    filter_complex.push_str(&timeline);
    
    // Add audio
    filter_complex.push_str(&format!("[{}:a]aformat=sample_fmts=fltp:sample_rates=44100:channel_layouts=stereo,atrim=0:{}[audio];", sorted_image_paths.len(), total_duration));
    
    // Combine video and audio
    filter_complex.push_str("[outv][audio]concat=n=1:v=1:a=1[outv_a];");

    // Add word highlighting
    let mut word_filter = String::new();
    for word in asr_data.words.iter() {
        let text = word.word.replace("'", "'\\\\\\''").replace(":", "\\:"); // Escape single quotes and colons
        word_filter.push_str(&format!(
            "drawtext=fontfile=/Users/aditya/Documents/OSS/zappush/shortpro/backend/public/Roboto-Bold.ttf:fontsize=120:fontcolor=white:box=1:boxcolor=black@0.5:boxborderw=10:x=(w-tw)/2:y=(h-th)/2:text='{}':enable='between(t,{},{})':alpha='if(between(t,{},{}),1,0)',",
            text, word.start, word.end, word.start, word.end
        ));
        
        // Highlight the current word
        word_filter.push_str(&format!(
            "drawtext=fontfile=/Users/aditya/Documents/OSS/zappush/shortpro/backend/public/Roboto-Bold.ttf:fontsize=120:fontcolor=yellow:box=1:boxcolor=black@0.5:boxborderw=10:x=(w-tw)/2:y=(h-th)/2:text='{}':enable='between(t,{},{})':alpha='if(between(t,{},{}),1,0)',",
            text, word.start, word.end, word.start, word.end
        ));
    }
    
    // Remove trailing comma and add to filter complex
    word_filter.pop();
    filter_complex.push_str(&format!("[outv_a]{}[output]", word_filter));

    ffmpeg_args.extend(vec!["-filter_complex".to_string(), filter_complex]);

    // Output mapping
    ffmpeg_args.extend(vec![
        "-map".to_string(), "[output]".to_string(),
        "-c:a".to_string(), "aac".to_string(),
        "-c:v".to_string(), "libx264".to_string(),
        "-preset".to_string(), "medium".to_string(),
        "-crf".to_string(), "23".to_string(),
        "-movflags".to_string(), "+faststart".to_string(),
        "-pix_fmt".to_string(), "yuv420p".to_string(),
    ]);

    // Add output file
    ffmpeg_args.push(output_file.to_string());

    println!("FFmpeg command: ffmpeg {}", ffmpeg_args.join(" "));

    // Run FFmpeg command
    println!("Starting FFmpeg process");
    let output = Command::new("ffmpeg")
        .args(&ffmpeg_args)
        .output()
        .context("Failed to execute FFmpeg command")?;

    if !output.status.success() {
        let error_msg = String::from_utf8_lossy(&output.stderr);
        println!("FFmpeg error: {}", error_msg);
        return Err(anyhow!("FFmpeg error: {}", error_msg));
    }

    let duration = start_time.elapsed();
    println!("Slideshow creation completed in {:.2} seconds", duration.as_secs_f64());

    Ok(())
}

fn create_subtitle_file(asr_data: &ASRData, output_file: &str) -> Result<()> {
    let mut content = String::new();
    for (i, word) in asr_data.words.iter().enumerate() {
        let start = format_time(word.start);
        let end = format_time(word.end);
        content.push_str(&format!("{}\n{} --> {}\n{}\n\n", i + 1, start, end, word.word));
    }
    fs::write(output_file, content).context("Failed to write subtitle file")?;
    Ok(())
}

fn format_time(seconds: f64) -> String {
    let hours = (seconds / 3600.0) as i32;
    let minutes = ((seconds % 3600.0) / 60.0) as i32;
    let secs = (seconds % 60.0) as i32;
    let millis = ((seconds - seconds.floor()) * 1000.0) as i32;
    format!("{:02}:{:02}:{:02},{:03}", hours, minutes, secs, millis)
}

#[tokio::main]
async fn main() {
    std::env::set_var("GOOGLE_APPLICATION_CREDENTIALS", "/Users/aditya/Documents/OSS/zappush/shortpro/backend/gcp_credentials.json");

    let create_slideshow = warp::post()
        .and(warp::path("create_slideshow"))
        .and(warp::body::json())
        .and_then(create_slideshow);

    println!("Starting server at http://127.0.0.1:8080");
    warp::serve(create_slideshow)
        .run(([127, 0, 0, 1], 8080))
        .await;
}