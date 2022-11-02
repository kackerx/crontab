package common

const (
    // 任务保存目录
    JOB_SAVE_DIR = "/cron/jobs/"

    // 任务强杀目录
    JOB_KILL_DIR = "/cron/killer/"

    // 锁
    JOB_LOCK_DIR = "/cron/lock/"
)

const (
    // 保存事件
    JOB_EVENT_SAVE = iota

    // 删除事件
    JOB_EVENT_DELETE

    // 强杀事件
    JOB_EVENT_KILL
)
