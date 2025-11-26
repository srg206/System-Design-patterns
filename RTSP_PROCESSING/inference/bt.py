import base64
from pathlib import Path
def main():
    p = Path('/Users/a1111/GO_PROJECTS/patterns/System-Design-patterns/RTSP_PROCESSING/inference/frame_000000.jpg')
    if not p.exists(): raise SystemExit(f'Файл не найден: {p}')
    b64 = base64.b64encode(p.read_bytes()).decode('ascii')
    out = p.with_suffix(p.suffix + '.b64.txt')
    out.write_text(b64)
    print(f'base64 записан в {out} ({len(b64)} символов)')
if __name__ == '__main__': main()