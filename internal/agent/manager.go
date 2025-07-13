package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Manager 负责启动/停止 Python agent 进程并提供简易聊天接口。
// 目前假设 agent/server.py 监听 127.0.0.1:8000。

var (
	mgr  *Manager
	once sync.Once
)

// Manager struct
//
// process: python 子进程句柄
// mu: 保证 Start/Stop 同步
// apiKey: 当前使用的 Key（用于是否需要重启）
// started: 标记子进程是否已成功启动

type Manager struct {
	process *exec.Cmd
	mu      sync.Mutex
	apiKey  string
	started bool
}

// Get 返回单例
func Get() *Manager {
	once.Do(func() { mgr = &Manager{} })
	return mgr
}

// Start 使用给定 apikey 启动（或重启）agent 进程。
// 若 apiKey 不变且进程存活，则直接返回。
func (m *Manager) Start(apiKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 若 key 相同且进程健康
	if apiKey == m.apiKey && m.started && m.process != nil && m.process.ProcessState == nil {
		return nil // 已在运行
	}

	// 如有旧进程先停止
	_ = m.stopLocked()

	// 设置 env
	env := os.Environ()
	env = append(env, fmt.Sprintf("OPENAI_API_KEY=%s", apiKey))
	// 可根据需要设置 OPENAI_MODEL 等

	cmd := exec.Command("python", "-u", "-m", "agent.server")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env

	if err := cmd.Start(); err != nil {
		return err
	}

	m.process = cmd
	m.apiKey = apiKey
	m.started = true

	// 后台等待进程退出，退出时清理状态
	go func() {
		_ = cmd.Wait()
		m.mu.Lock()
		m.started = false
		m.process = nil
		m.mu.Unlock()
	}()

	// 简单等待 1s 以确保 HTTP 服务就绪
	time.Sleep(time.Second)
	return nil
}

// Stop 停止子进程
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	_ = m.stopLocked()
}

func (m *Manager) stopLocked() error {
	if m.process == nil || m.process.Process == nil {
		return nil
	}
	_ = m.process.Process.Kill()
	m.process = nil
	m.started = false
	return nil
}

// SendMessage 向本地 agent 发送聊天请求
func (m *Manager) SendMessage(message string) (string, error) {
	if !m.started {
		return "", fmt.Errorf("AI 服务未启动")
	}
	payload := map[string]string{"message": message}
	b, _ := json.Marshal(payload)
	resp, err := http.Post("http://127.0.0.1:8000/chat", "application/json", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var data struct {
		Reply string `json:"reply"`
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI 服务错误: %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	return data.Reply, nil
}
