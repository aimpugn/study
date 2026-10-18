/*
 * 탐색용 코드입니다(2026-10-13, 작업대 관측 재설계의 P15~P20 시험).
 * 시험 결과에 따라 실습 이미지의 빌드 단계로 옮기거나 지웁니다.
 *
 * gcc -finstrument-functions로 컴파일한 PostgreSQL은 모든 함수의 입구와 출구에서
 * __cyg_profile_func_enter와 __cyg_profile_func_exit를 부릅니다. 이 파일이 그 두 함수를 정의합니다.
 * 사건마다 고정 크기 레코드 하나를 프로세스별 파일에 write() 한 번으로 남깁니다.
 *
 * 설정: 환경 변수 PGTRACE_DIR  기록 파일을 둘 디렉터리. 없으면 아무것도 기록하지 않습니다.
 * 출력: <PGTRACE_DIR>/<pid>.bin. 레코드는 8바이트 정수 다섯(kind, fn, site, ns, frame)이고 리틀 엔디언입니다.
 *       kind 0 = 파일 머리(fn = 실행 때 main의 주소, site = pid)
 *       kind 1 = 함수 입구, kind 2 = 함수 출구(fn = 함수 주소, site = 부른 자리 주소)
 *       kind 3 = 기록 도중 끼어든 신호 처리기 때문에 버린 사건 수(fn = 개수)
 *       ns는 CLOCK_REALTIME 나노초입니다. strace -ttt와 같은 시계라서 두 기록을 시각으로 맞출 수 있습니다.
 *       frame은 기록된 함수의 프레임 주소입니다. elog(ERROR)의 longjmp는 중간 함수들의 출구를 건너뛰므로,
 *       해석기는 새 입구의 frame보다 낮은 주소의 칸(스택은 아래로 자람)이 이미 사라졌다고 판단합니다.
 *       이 빌드는 -fno-omit-frame-pointer라서 프레임 주소를 믿을 수 있습니다.
 *
 * 컴파일할 때 __builtin_frame_address(1)에 대한 -Wframe-address 경고가 하나 나옵니다. 프레임 포인터를 남기는
 * 이 빌드(-fno-omit-frame-pointer)에서는 의도한 사용이므로 고치지 않습니다.
 *
 * 이 파일의 함수는 모두 no_instrument_function이고 libc만 부릅니다.
 * PostgreSQL 함수를 부르면 기록 호출이 다시 기록 호출을 불러 끝나지 않습니다.
 */
#include <errno.h>
#include <fcntl.h>
#include <pthread.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include <unistd.h>

#define NOI __attribute__((no_instrument_function))

/* 레코드 하나는 40바이트이고, 해석기(Go)가 같은 배치로 읽습니다. */
struct pgtrace_rec
{
	uint64_t	kind;
	uint64_t	fn;
	uint64_t	site;
	uint64_t	ns;
	uint64_t	frame;
};

extern int	main(int argc, char *argv[]);

void		__cyg_profile_func_enter(void *fn, void *site) NOI;
void		__cyg_profile_func_exit(void *fn, void *site) NOI;

/* -1은 아직 열지 않음, -2는 이 프로세스에서 기록하지 않음입니다. */
static int	trace_fd = -1;
static int	busy;
static uint64_t dropped;

/* 사건 시각을 나노초로 돌려줍니다. vDSO를 거치므로 시스템 콜이 아닙니다. */
NOI static uint64_t
now_ns(void)
{
	struct timespec ts;

	clock_gettime(CLOCK_REALTIME, &ts);
	return (uint64_t) ts.tv_sec * 1000000000u + (uint64_t) ts.tv_nsec;
}

/* 이 프로세스의 기록 파일을 열고, 주소 해석에 쓸 main의 실행 때 주소를 머리에 적습니다. */
NOI static void
open_trace(void)
{
	const char *dir = getenv("PGTRACE_DIR");
	char		path[512];
	struct pgtrace_rec head;

	if (dir == NULL)
	{
		trace_fd = -2;
		return;
	}
	snprintf(path, sizeof path, "%s/%d.bin", dir, (int) getpid());
	trace_fd = open(path, O_WRONLY | O_CREAT | O_APPEND | O_CLOEXEC, 0600);
	if (trace_fd < 0)
	{
		trace_fd = -2;
		return;
	}
	head.kind = 0;
	head.fn = (uint64_t) (uintptr_t) &main;
	head.site = (uint64_t) getpid();
	head.ns = now_ns();
	head.frame = 0;
	(void) write(trace_fd, &head, sizeof head);
}

/*
 * fork한 자식이 부모의 파일에 이어 쓰지 않도록, 자식에서는 물려받은 파일을 닫고 다음 사건에서 자기 파일을 엽니다.
 * 버퍼를 두지 않으므로 부모의 미기록분을 물려받는 일은 없습니다.
 */
NOI static void
after_fork_child(void)
{
	if (trace_fd >= 0)
		close(trace_fd);
	trace_fd = -1;
	busy = 0;
	dropped = 0;
}

NOI __attribute__((constructor)) static void
register_fork_handler(void)
{
	pthread_atfork(NULL, NULL, after_fork_child);
}

/*
 * 사건 하나를 기록합니다. 기록 도중 신호 처리기가 끼어들어 다시 들어오면 그 사건은 버리고 개수만 셉니다.
 * 처리기는 기록이 다시 이어지기 전에 끝나므로, 처리기 안의 입구와 출구는 짝으로 함께 버려집니다.
 * PostgreSQL은 시스템 콜 직후 errno를 읽으므로, 기록이 errno를 바꾸지 않게 보존합니다.
 */
NOI static void
emit(uint64_t kind, void *fn, void *site, void *frame)
{
	int			saved_errno = errno;
	struct pgtrace_rec rec;

	if (busy)
	{
		dropped++;
		return;
	}
	busy = 1;
	if (trace_fd == -1)
		open_trace();
	if (trace_fd >= 0)
	{
		if (dropped > 0)
		{
			rec.kind = 3;
			rec.fn = dropped;
			rec.site = 0;
			rec.ns = now_ns();
			rec.frame = 0;
			(void) write(trace_fd, &rec, sizeof rec);
			dropped = 0;
		}
		rec.kind = kind;
		rec.fn = (uint64_t) (uintptr_t) fn;
		rec.site = (uint64_t) (uintptr_t) site;
		rec.ns = now_ns();
		rec.frame = (uint64_t) (uintptr_t) frame;
		(void) write(trace_fd, &rec, sizeof rec);
	}
	busy = 0;
	errno = saved_errno;
}

/*
 * __builtin_frame_address(1)은 이 함수를 부른 함수, 곧 기록되는 함수의 프레임 주소입니다.
 * 기록 함수를 다른 함수로 감싸면 단계가 하나 밀리므로 여기서 바로 구합니다.
 */
void
__cyg_profile_func_enter(void *fn, void *site)
{
	emit(1, fn, site, __builtin_frame_address(1));
}

void
__cyg_profile_func_exit(void *fn, void *site)
{
	emit(2, fn, site, __builtin_frame_address(1));
}
