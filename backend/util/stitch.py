import os
import cv2
import pysrt
import random
import numpy as np
from pathlib import Path
from moviepy.editor import *

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

def create_smooth_shake_transition(clip1, clip2, video_size, duration=0.5, shake_intensity=3):
    w, h = video_size
    
    def make_frame(t):
        progress = t / duration
        
        # Get frames from both clips
        frame1 = clip1.get_frame(clip1.duration * (1 - 0.1 * (1 - progress)))
        frame2 = clip2.get_frame(clip2.duration * (0.1 * progress))
        
        # Apply shake effect
        shake_amount = shake_intensity * np.sin(progress * np.pi)  # Smooth shake curve
        dx = int(shake_amount * (np.random.random() - 0.5))
        dy = int(shake_amount * (np.random.random() - 0.5))
        
        # Apply shake to both frames
        M1 = np.float32([[1, 0, dx], [0, 1, dy]])
        M2 = np.float32([[1, 0, -dx], [0, 1, -dy]])  # Inverse shake for second frame
        frame1 = cv2.warpAffine(frame1, M1, (w, h), borderMode=cv2.BORDER_REFLECT)
        frame2 = cv2.warpAffine(frame2, M2, (w, h), borderMode=cv2.BORDER_REFLECT)
        
        # Smooth crossfade
        weight = np.sin(progress * np.pi / 2)**2  # Smooth S-curve for transition
        frame = cv2.addWeighted(frame1, 1-weight, frame2, weight, 0)
        
        return frame

    return VideoClip(make_frame, duration=duration)

def create_slideshow_with_subtitles(image_paths, srt_file, audio_file, output_file, whoosh_file):
    subs = pysrt.open(srt_file)
    effects = ["ken_burns", "zoom_out", "pan"]
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
        
        # Ensure the clip is in RGB format
        if clip.get_frame(0).shape[2] == 4:  # If it has an alpha channel
            clip = clip.set_mask(None)  # Remove the alpha channel
        
        clips.append(clip)
        
        # Add transition after each clip (except the last one)
        # if i < len(image_paths) - 1:
        #     next_clip = create_varied_clip(image_paths[i+1], duration, random.choice(effects))
        #     visual_transition = create_smooth_shake_transition(clip, next_clip, video_size, transition_duration)
        #     transition_with_audio = visual_transition.set_audio(whoosh_audio)
        #     transitions.append(transition_with_audio)
    
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

if __name__ == "__main__":
    image_paths = sorted([f'/tmp/{f}' for f in os.listdir('/tmp') if f.startswith('image_')])
    srt_file = '/tmp/subtitles.srt'
    audio_file = '/tmp/full_audio.mp3'
    whoosh_file = "/Users/aditya/Documents/OSS/zappush/shortpro/backend/public/15. Whoosh Swoosh.wav"
    output_file = '/tmp/output.mp4'
    create_slideshow_with_subtitles(image_paths, srt_file, audio_file, output_file, whoosh_file)