import os
from pathlib import Path
import tempfile
from faster_whisper import WhisperModel
from moviepy.video.VideoClip import ColorClip
from moviepy.editor import *
from flask import Flask, request, jsonify

from textwrap import wrap

# Set up MoviePy configuration
import moviepy.config as cfg
cfg.IMAGEMAGICK_BINARY = "/opt/homebrew/bin/convert"
import os
os.environ["IMAGEIO_FFMPEG_EXE"] = "/opt/homebrew/bin/ffmpeg"

# Set up MPS for GPU acceleration
os.environ["CUDA_VISIBLE_DEVICES"] = ""
os.environ["PYTORCH_ENABLE_MPS_FALLBACK"] = "1"

app = Flask(__name__)

image_paths = sorted([f'/tmp/{f}' for f in os.listdir('/tmp') if f.startswith('image_')])
audio_file = '/tmp/full_audio.mp3'
output_file = '/tmp/output.mp4'

REEL_ASPECT_RATIO = 9 / 16  # Typical aspect ratio for reels (1080x1920)

def generate_asr_data(audio_file):
    model = WhisperModel("base", device="cpu", compute_type="int8")
    segments, info = model.transcribe(audio_file, word_timestamps=True)

    asr_data = {
        "sentences": [],
        "words": []
    }

    for segment in segments:
        asr_data["sentences"].append({
            "start": segment.start,
            "end": segment.end,
            "text": segment.text
        })
        for word in segment.words:
            asr_data["words"].append({
                "start": word.start,
                "end": word.end,
                "word": word.word
            })

    return asr_data

# def create_subtitle_clip(words, video_size, max_duration):
#     w, h = video_size
#     font_path = Path('/Users/aditya/Documents/OSS/zappush/shortpro/backend/public/Roboto-Bold.ttf')
#     font = str(font_path) if font_path.exists() else 'Arial'
#     fontsize = h // 20

#     word_clips = []
#     y_pos = h // 2

#     for i in range(0, len(words), 3):
#         current_words = words[i:i+3]
#         text = " ".join([word['word'] for word in current_words])
        
#         clip = TextClip(text, font=font, fontsize=fontsize, color='white',
#                         stroke_color='black', stroke_width=2, method='label')
        
#         clip = clip.set_position(('center', y_pos))
        
#         start_time = current_words[0]['start']
#         end_time = current_words[-1]['end']
        
#         clip = clip.set_start(start_time).set_duration(end_time - start_time)
        
#         word_clips.append(clip)

#     subtitle_clip = CompositeVideoClip(word_clips, size=video_size).set_duration(max_duration)
#     return subtitle_clip

def create_subtitle_clip(words, video_size, max_duration):
    w, h = video_size
    font_path = Path('/Users/aditya/Documents/OSS/zappush/shortpro/backend/public/Roboto-Bold.ttf')
    font = str(font_path) if font_path.exists() else 'Arial'
    fontsize = h // 20

    word_clips = []
    y_pos = h // 2

    def create_text_clip(text, start, end):
        # Wrap text if it's too wide
        wrapped_text = '\n'.join(wrap(text, width=20))  # Adjust width as needed
        
        clip = TextClip(wrapped_text, font=font, fontsize=fontsize, color='white',
                        stroke_color='black', stroke_width=3, method='label')
        clip = clip.set_position(('center', 'center'))
        return clip.set_start(start).set_duration(end - start)

    current_text = []
    current_start = words[0]['start']

    for i, word in enumerate(words):
        current_text.append(word['word'])
        
        # Check if we have 3 words or if it's the last word
        if len(current_text) == 3 or i == len(words) - 1:
            # Check if the next word starts a new sentence
            if i + 1 < len(words) and words[i+1]['word'].istitle():
                word_clips.append(create_text_clip(' '.join(current_text), current_start, word['end']))
                current_text = []
                current_start = words[i+1]['start'] if i + 1 < len(words) else word['end']
            elif '.' in word['word'] or '?' in word['word'] or '!' in word['word']:
                word_clips.append(create_text_clip(' '.join(current_text), current_start, word['end']))
                current_text = []
                current_start = words[i+1]['start'] if i + 1 < len(words) else word['end']
            elif len(current_text) == 3:
                word_clips.append(create_text_clip(' '.join(current_text), current_start, word['end']))
                current_text = []
                current_start = words[i+1]['start'] if i + 1 < len(words) else word['end']

    subtitle_clip = CompositeVideoClip(word_clips, size=video_size, ).set_duration(max_duration)
    return subtitle_clip

def create_slideshow_with_subtitles(image_paths, asr_data, audio_file, output_file):
    sentences = asr_data['sentences']
    words = asr_data['words']
    
    audio = AudioFileClip(audio_file)
    total_duration = audio.duration

    reel_width = 1080
    reel_height = int(reel_width / REEL_ASPECT_RATIO)
    reel_size = (reel_width, reel_height)

    clips = []
    current_time = 0
    for i, sentence in enumerate(sentences):
        start, end = sentence['start'], sentence['end']
        
        if start > current_time and clips:
            gap_duration = start - current_time
            gap_clip = clips[-1].copy().set_duration(gap_duration).set_start(current_time)
            clips.append(gap_clip)
        
        duration = end - start
        img = ImageClip(image_paths[min(i, len(image_paths)-1)])
        img_aspect_ratio = img.w / img.h
        if img_aspect_ratio > REEL_ASPECT_RATIO:
            new_height = reel_height
            new_width = int(new_height * img_aspect_ratio)
        else:
            new_width = reel_width
            new_height = int(new_width / img_aspect_ratio)
        
        img_resized = img.resize(height=new_height, width=new_width).set_position(('center', 'center'))
        
        bg_clip = ColorClip(size=reel_size, color=(0, 0, 0))
        
        clip = CompositeVideoClip([bg_clip, img_resized]).set_duration(duration).set_start(start)
        clips.append(clip)
        current_time = end
    
    if current_time < total_duration:
        final_gap = total_duration - current_time
        final_clip = clips[-1].copy().set_duration(final_gap).set_start(current_time)
        clips.append(final_clip)

    video = CompositeVideoClip(clips, size=reel_size, )
    
    subtitle_clip = create_subtitle_clip(words, reel_size, total_duration)
    final_video = CompositeVideoClip([video, subtitle_clip], )
    
    final_video = final_video.set_audio(audio)
    final_video = final_video.set_duration(total_duration)
    
    final_video.write_videofile(output_file, fps=30, audio_codec='aac', codec='h264_videotoolbox')

@app.route('/generate_asr', methods=['POST'])
def generate_asr():
    if 'audio' not in request.files:
        return jsonify({"error": "No audio file provided"}), 400
    
    audio_file = request.files['audio']
    
    if audio_file.filename == '':
        return jsonify({"error": "No selected file"}), 400
    
    if audio_file and audio_file.filename.lower().endswith(('.mp3', '.wav', '.flac')):
        with tempfile.NamedTemporaryFile(delete=False, suffix="." + audio_file.filename.split('.')[-1]) as temp_audio:
            audio_file.save(temp_audio.name)
            asr_data = generate_asr_data(temp_audio.name)
        
        return jsonify(asr_data)
    else:
        return jsonify({"error": "Invalid file format. Please upload an MP3, WAV, or FLAC file."}), 400

@app.route('/create_slideshow', methods=['POST'])
def api_create_slideshow():
    data = request.json
    
    asr_data = data.get('asr_data')

    print(image_paths)
    print(asr_data)
    print(audio_file)
    print(output_file)

    create_slideshow_with_subtitles(image_paths, asr_data, audio_file, output_file)
    return jsonify({"message": "Slideshow created successfully", "output_file": output_file}), 200

if __name__ == "__main__":
    app.run(debug=True)