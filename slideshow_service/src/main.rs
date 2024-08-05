use warp::Filter;
use serde::{Deserialize, Serialize};
use std::convert::Infallible;
use std::process::Command;
use std::time::Instant;
use log::{debug, info, error};


use std::fs;
use std::path::{PathBuf, Path};

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
    asr_data: ASRData,
}

#[derive(Debug, Serialize)]
struct CreateSlideshowResponse {
    message: String,
    output_file: String,
}

async fn create_slideshow(req: CreateSlideshowRequest) -> Result<impl warp::Reply, Infallible> {
    let image_paths: Vec<PathBuf> = fs::read_dir("/tmp")
        .unwrap()
        .filter_map(|entry| {
            let entry = entry.unwrap();
            let path = entry.path();
            if path.is_file() && path.file_name().unwrap().to_str().unwrap().starts_with("image_") {
                Some(path)
            } else {
                None
            }
        })
        .collect();

    let audio_file = "/tmp/full_audio.mp3";
    let output_file = "/tmp/output_rust.mp4";

    println!("Image paths: {:?}", image_paths);
    println!("ASR data: {:?}", req.asr_data);
    println!("Audio file: {}", audio_file);
    println!("Output file: {}", output_file);

    match create_slideshow_with_subtitles(&image_paths, &req.asr_data, audio_file, output_file) {
        Ok(_) => {
            let response = CreateSlideshowResponse {
                message: "Slideshow created successfully".to_string(),
                output_file: output_file.to_string(),
            };
            Ok(warp::reply::json(&response))
        },
        Err(e) => {
            let error_response = CreateSlideshowResponse {
                message: format!("Error creating slideshow: {}", e),
                output_file: "".to_string(),
            };
            Ok(warp::reply::json(&error_response))
        }
    }
}

fn create_slideshow_with_subtitles(
    image_paths: &[PathBuf],
    asr_data: &ASRData,
    audio_file: &str,
    output_file: &str,
) -> Result<(), Box<dyn std::error::Error>> {
    let start_time = Instant::now();

    // Ensure the output directory exists
    if let Some(parent) = Path::new(output_file).parent() {
        fs::create_dir_all(parent)?;
        debug!("Created output directory: {:?}", parent);
    }

    // Create subtitle file
    create_subtitle_file(asr_data, "/tmp/subtitles.srt")?;
    info!("Created subtitle file");

    // Prepare FFmpeg command
    let mut ffmpeg_args = vec![
        "-y",  // Overwrite output file if it exists
        "-i", audio_file,
    ];

    // Add input images
    for path in image_paths {
        ffmpeg_args.extend_from_slice(&["-loop", "1", "-t", "5", "-i", path.to_str().unwrap()]);
    }

    // Create filter complex
    let mut filter_complex = String::new();
    for i in 0..image_paths.len() {
        filter_complex.push_str(&format!("[{}:v]scale={}:{}:force_original_aspect_ratio=decrease,pad={}:{}:(ow-iw)/2:(oh-ih)/2[v{}];", 
            i + 1, REEL_WIDTH, REEL_HEIGHT, REEL_WIDTH, REEL_HEIGHT, i));
    }
    filter_complex.push_str(&format!("{}concat=n={}:v=1:a=0[outv];", 
        (0..image_paths.len()).map(|i| format!("[v{}]", i)).collect::<Vec<_>>().join(""), 
        image_paths.len()));
    
    // Add subtitles to the complex filter
    filter_complex.push_str("[outv]subtitles=/tmp/subtitles.srt:force_style='FontSize=24,Alignment=10'[outv_sub]");

    ffmpeg_args.extend_from_slice(&["-filter_complex", &filter_complex]);

    // Output mapping
    ffmpeg_args.extend_from_slice(&[
        "-map", "[outv_sub]",
        "-map", "0:a",
        "-c:a", "aac",
        "-c:v", "libx264",
        "-preset", "medium",
        "-crf", "23",
        "-shortest",
    ]);

    // Add output file
    ffmpeg_args.push(output_file);

    debug!("FFmpeg command: ffmpeg {}", ffmpeg_args.join(" "));

    // Run FFmpeg command
    info!("Starting FFmpeg process");
    let output = Command::new("ffmpeg")
        .args(&ffmpeg_args)
        .output()?;

    if !output.status.success() {
        let error_msg = String::from_utf8_lossy(&output.stderr);
        error!("FFmpeg error: {}", error_msg);
        return Err(Box::new(std::io::Error::new(
            std::io::ErrorKind::Other,
            format!("FFmpeg error: {}", error_msg),
        )));
    }

    let duration = start_time.elapsed();
    info!("Slideshow creation completed in {:.2} seconds", duration.as_secs_f64());

    Ok(())
}


fn create_subtitle_file(asr_data: &ASRData, output_file: &str) -> Result<(), Box<dyn std::error::Error>> {
    let mut content = String::new();
    for (i, word) in asr_data.words.iter().enumerate() {
        let start = format_time(word.start);
        let end = format_time(word.end);
        content.push_str(&format!("{}\n{} --> {}\n{}\n\n", i + 1, start, end, word.word));
    }
    fs::write(output_file, content)?;
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
    let create_slideshow = warp::post()
        .and(warp::path("create_slideshow"))
        .and(warp::body::content_length_limit(1024 * 1024 * 50))
        .and(warp::body::json())
        .and_then(create_slideshow);

    println!("Starting server at http://127.0.0.1:8080");
    warp::serve(create_slideshow)
        .run(([127, 0, 0, 1], 8080))
        .await;
}