/* opl_render — 離線把一條 CMF 變體軌用 dq3_opl2 render 成 wav,驗證 OPL2 音色。
 * 編譯(docker 內):cc -O2 -I dq3_remake/src tools/opl_render.c \
 *                    dq3_remake/src/dq3_opl2.c dq3_remake/src/dq3_cmf.c -lm -o /tmp/opl_render
 * 用法:opl_render track.bin out.wav [seconds] [tick_ms]
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "dq3_opl2.h"
#include "dq3_cmf.h"

#define SR 44100

static void wr16(FILE *f, int v){ fputc(v&0xff,f); fputc((v>>8)&0xff,f); }
static void wr32(FILE *f, long v){ fputc(v&0xff,f); fputc((v>>8)&0xff,f); fputc((v>>16)&0xff,f); fputc((v>>24)&0xff,f); }

static void write_wav(const char *path, const short *pcm, int n){
    FILE *f = fopen(path,"wb");
    long data = (long)n*2;
    fwrite("RIFF",1,4,f); wr32(f,36+data); fwrite("WAVE",1,4,f);
    fwrite("fmt ",1,4,f); wr32(f,16); wr16(f,1); wr16(f,1);
    wr32(f,SR); wr32(f,SR*2); wr16(f,2); wr16(f,16);
    fwrite("data",1,4,f); wr32(f,data);
    fwrite(pcm,2,n,f); fclose(f);
}

int main(int argc, char **argv){
    const char *src = argc>1?argv[1]:"work/music/track_00.bin";
    const char *out = argc>2?argv[2]:"work/music/opl_00.wav";
    double secs = argc>3?atof(argv[3]):25.0;
    double tick = argc>4?atof(argv[4]):13.0;
    FILE *f = fopen(src,"rb"); long len; unsigned char *buf;
    dq3_opl2 *opl; dq3_cmf *cmf; int total, done; short *pcm;
    if(!f){ fprintf(stderr,"open %s fail\n",src); return 1; }
    fseek(f,0,SEEK_END); len=ftell(f); fseek(f,0,SEEK_SET);
    buf=malloc(len); fread(buf,1,len,f); fclose(f);

    opl = dq3_opl2_new(SR);
    cmf = dq3_cmf_load(buf, (int)len, opl, SR, tick);
    if(!cmf){ fprintf(stderr,"cmf load fail (no events?)\n"); return 2; }
    printf("events=%d est_len=%.1fs  tick=%.1fms\n",
           dq3_cmf_event_count(cmf), dq3_cmf_length_sec(cmf), tick);

    total = (int)(secs*SR);
    pcm = calloc(total, sizeof(short));
    done = 0;
    while(done < total){
        int chunk = 4096; if(chunk>total-done) chunk=total-done;
        dq3_cmf_render(cmf, pcm+done, chunk, 1 /*loop*/);
        done += chunk;
    }
    write_wav(out, pcm, total);
    printf("wrote %s (%.1fs)\n", out, secs);
    dq3_cmf_free(cmf); dq3_opl2_free(opl); free(pcm); free(buf);
    return 0;
}
