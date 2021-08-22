package internal

import (
	pkg2 "crontab/master/pkg"
	"crontab/pkg"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"time"
)

func routers() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/job/save", SaveJob)
	mux.HandleFunc("/job/delete", DeleteJob)
	mux.HandleFunc("/job/list", ListJob)
	mux.HandleFunc("/job/kill", KillJob)

	return mux
}

func Run() error {
	server := &http.Server{
		ReadTimeout:  time.Duration(pkg2.G_config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(pkg2.G_config.WriteTimeout) * time.Second,
		Handler:      routers(),
	}

	listen, err := net.Listen("tcp", ":"+strconv.Itoa(pkg2.G_config.ApiPort))
	if err != nil {
		return err
	}

	return server.Serve(listen)
}

func SaveJob(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	err := r.ParseForm()
	if err != nil {
		_, _ = w.Write(pkg.SystemError("请求参数错误").ToJson())
		return
	}

	postJob := r.PostForm.Get("job")
	job := &pkg.Job{}
	err = json.Unmarshal([]byte(postJob), job)
	if err != nil {
		_, _ = w.Write(pkg.Fail("请求参数格式错误").ToJson())
		return
	}

	preJob, err := pkg2.G_jobMgr.SaveJob("/cron/jobs", job)
	if err != nil {
		_, _ = w.Write(pkg.Fail(err.Error()).ToJson())
		return
	}

	_, _ = w.Write(pkg.Success().WithData(preJob).ToJson())
}

func DeleteJob(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	err := r.ParseForm()
	if err != nil {
		_, _ = w.Write(pkg.SystemError("请求参数错误").ToJson())
		return
	}

	jobName := r.Form.Get("job")
	job, err := pkg2.G_jobMgr.DeleteJob("/cron/jobs/", jobName)
	if err != nil {
		_, _ = w.Write(pkg.Fail("删除异常:" + err.Error()).ToJson())
	}

	_, _ = w.Write(pkg.Success().WithData(job).ToJson())
}

func ListJob(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	err := r.ParseForm()
	if err != nil {
		_, _ = w.Write(pkg.SystemError("系统异常").ToJson())
		return
	}

	namespace := r.Form.Get("namespace")
	jobs, err := pkg2.G_jobMgr.ListJobs(namespace)
	if err != nil {
		_, _ = w.Write(pkg.Fail("获取失败:" + err.Error()).ToJson())
		return
	}

	_, _ = w.Write(pkg.Success().WithData(jobs).ToJson())
}

func KillJob(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "applicaion/json")

	err := r.ParseForm()
	if err != nil {
		_, _ = w.Write(pkg.SystemError("系统错误").ToJson())
		return
	}

	jobName := r.Form.Get("job")
	err = pkg2.G_jobMgr.KillJob(jobName)
	if err != nil {
		_, _ = w.Write(pkg.Fail("杀死任务：" + jobName + "失败, " + err.Error()).ToJson())
		return
	}

	_, _ = w.Write(pkg.Success().ToJson())
}
