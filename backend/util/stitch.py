import os
import random
from pathlib import Path
import tempfile
import whisper

import numpy as np
import pysrt
from moviepy.editor import *
from flask import Flask, request, jsonify, send_file

app = Flask(__name__)

image_paths = sorted([f'/tmp/{f}' for f in os.listdir('/tmp') if f.startswith('image_')])
srt_file = '/tmp/subtitles.srt'
audio_file = '/tmp/full_audio.mp3'
whoosh_file = "/Users/aditya/Documents/OSS/zappush/shortpro/backend/public/15. Whoosh Swoosh.wav"
output_file = '/tmp/output.mp4'

def time_to_seconds(time_obj):
    return time_obj.hours * 3600 + time_obj.minutes * 60 + time_obj.seconds + time_obj.milliseconds / 1000

def create_varied_clip(img_path, duration, effect):
    img = ImageClip(img_path).set_duration(duration)
    
    # Ensure the image is in RGB format
    if img.img.shape[2] == 4:  # If it has an alpha channel
        img = img.set_mask(None)  # Remove the alpha channel
    
    w, h = img.size
     
    if effect == "ken_burns":
        def ken_burns(t):
            zoom = 1 + t/duration * 0.1
            new_w, new_h = int(w * zoom), int(h * zoom)
            x1 = int((new_w - w) / 2)
            y1 = int((new_h - h) / 2)
            return img.resize((new_w, new_h)).crop(x1=x1, y1=y1, x2=x1+w, y2=y1+h).get_frame(t)
        return VideoClip(make_frame=ken_burns, duration=duration)
    
    elif effect == "pan":
        zoom_factor = 1.1  # Zoom in by 10%
        def pan(t):
            new_w, new_h = int(w * zoom_factor), int(h * zoom_factor)
            max_pan = new_w - w
            x_offset = int((t / duration) * max_pan)
            return (img.resize((new_w, new_h))
                    .crop(x1=x_offset, y1=0, x2=x_offset+w, y2=h)
                    .get_frame(t))
        return VideoClip(make_frame=pan, duration=duration)
    
    elif effect == "zoom_out":
        def zoom_out(t):
            zoom = 1.1 - t/duration * 0.1  # Start at 1.1x zoom and end at 1x
            new_w, new_h = int(w * zoom), int(h * zoom)
            x1 = int((new_w - w) / 2)
            y1 = int((new_h - h) / 2)
            return img.resize((new_w, new_h)).crop(x1=x1, y1=y1, x2=x1+w, y2=y1+h).get_frame(t)
        return VideoClip(make_frame=zoom_out, duration=duration)

    else:  # Default to static image
        return img

def create_subtitle_clips(subs, video_size):
    w, h = video_size
    font_path = Path('/Users/aditya/Documents/OSS/zappush/shortpro/backend/public/Roboto-Bold.ttf')
    
    if not font_path.exists():
        print(f"Warning: Font file not found at {font_path}. Using default font.")
        font_path = None
    
    return [((time_to_seconds(sub.start), time_to_seconds(sub.end)),
             TextClip(sub.text, 
                      font=str(font_path) if font_path else 'Arial-Bold',
                      fontsize=h//10,
                      color='white',
                      stroke_color='black',
                      stroke_width=4,
                      method='caption',
                      size=(w*0.9, None),
                      align='center')
             .set_position(('center', 'center')))
            for sub in subs]

def create_slideshow_with_subtitles(image_paths, srt_file, audio_file, output_file, whoosh_file):
    subs = pysrt.open(srt_file)
    # effects = ["ken_burns", "zoom_out", "pan", "static"]
    effects = [ "static" ] # for now
    clips = []
    transitions = []
    
    first_img = ImageClip(image_paths[0])
    video_size = first_img.size
    transition_duration = 0.5  # Duration of the transition in seconds
    
    whoosh_audio = AudioFileClip(whoosh_file).set_duration(transition_duration)
    
    for i, sub in enumerate(subs):
        if i >= len(image_paths):
            break
        duration = time_to_seconds(sub.end) - time_to_seconds(sub.start)
        effect = random.choice(effects)
        clip = create_varied_clip(image_paths[i], duration, effect)
        
        # Ensure all clips have the same size and format
        clip = clip.resize(video_size)
        if clip.get_frame(0).dtype != np.uint8:
            clip = clip.set_position((0, 0))
        
        clips.append(clip)
    
    # Combine clips with transitions
    final_clips = []
    for i, clip in enumerate(clips):
        final_clips.append(clip)
        if i < len(transitions):
            final_clips.append(transitions[i])
    
    video = concatenate_videoclips(final_clips)
    subtitles = create_subtitle_clips(subs[:len(clips)], video_size)
    subtitles_clip = CompositeVideoClip([video] + [clip.set_start(t[0]).set_end(t[1]) for t, clip in subtitles])
    
    audio = AudioFileClip(audio_file)
    if subtitles_clip.duration > audio.duration:
        subtitles_clip = subtitles_clip.set_duration(audio.duration)
    else:
        audio = audio.subclip(0, subtitles_clip.duration)
    
    final_video = subtitles_clip.set_audio(audio)
    
    # Print the shape of the first frame to check for any issues
    print(f"Final video shape: {final_video.get_frame(0).shape}")
    
    final_video.write_videofile(output_file, fps=60, audio_codec='aac')
    
def generate_word_level_srt(audio_file):
    model = whisper.load_model("base")
    result = model.transcribe(audio_file, word_timestamps=True)
    
    srt_content = []
    for i, segment in enumerate(result["segments"], start=1):
        for j, word in enumerate(segment["words"], start=1):
            start = word["start"]
            end = word["end"]
            text = word["word"]
            
            start_time = f"{int(start // 3600):02d}:{int((start % 3600) // 60):02d}:{int(start % 60):02d},{int((start % 1) * 1000):03d}"
            end_time = f"{int(end // 3600):02d}:{int((end % 3600) // 60):02d}:{int(end % 60):02d},{int((end % 1) * 1000):03d}"
            
            srt_content.append(f"{i}.{j}\n{start_time} --> {end_time}\n{text}\n\n")
    
    return "".join(srt_content)

@app.route('/generate_srt', methods=['POST'])
def generate_srt():
    if 'audio' not in request.files:
        return jsonify({"error": "No audio file provided"}), 400
    
    audio_file = request.files['audio']
    
    if audio_file.filename == '':
        return jsonify({"error": "No selected file"}), 400
    
    if audio_file and audio_file.filename.endswith('.mp3'):
        with tempfile.NamedTemporaryFile(delete=False, suffix=".mp3") as temp_audio:
            audio_file.save(temp_audio.name)
            srt_content = generate_word_level_srt(temp_audio.name)
        
        with tempfile.NamedTemporaryFile(delete=False, suffix=".srt", mode="w", encoding="utf-8") as temp_srt:
            temp_srt.write(srt_content)
        
        return send_file(temp_srt.name, as_attachment=True, download_name="subtitles.srt")
    else:
        return jsonify({"error": "Invalid file format. Please upload an MP3 file."}), 400
 
@app.route('/create_slideshow', methods=['POST'])
def api_create_slideshow():
    data = request.json
    
    # image_paths = data.get('image_paths')        
    # srt_file = data.get('srt_file')
    # audio_file = data.get('audio_file')
    # whoosh_file = data.get('whoosh_file')
    # output_file = data.get('output_file')

    if not all([image_paths, srt_file, audio_file, whoosh_file, output_file]):
        return jsonify({"error": "Missing required parameters"}), 400

    try:
        create_slideshow_with_subtitles(image_paths, srt_file, audio_file, output_file, whoosh_file)
        return jsonify({"message": "Slideshow created successfully", "output_file": output_file}), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 500

if __name__ == "__main__":
    app.run(debug=True)