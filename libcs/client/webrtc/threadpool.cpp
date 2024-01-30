// Copyright (c) 2022 Institute of Software, Chinese Academy of Sciences (ISCAS)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#include <atomic>
#include <mutex>
#include <shared_mutex>
#include <vector>

#include <rtc_base/thread.h>

#include "threadpool.h"

// 实现 webrtc 的线程池管理，如果每个连接都分配三个线程的话，
// 在高并发的情况下性能较差。想象一个圆盘，盘面上有很多个扇形，
// 每个扇形代表一个线程，这些扇形刚开始是白色的，表示还没有初始化。
// 线程初始化之后扇形的颜色就变成了黑色，然后圆盘会转动起来，
// 越来越多的扇形变成了黑色。
class ThreadPool {
  public:
    ThreadPool(uint32_t threadNum) {
        threads = std::vector<std::unique_ptr<rtc::Thread>>(threadNum);
        socketThreads = std::vector<std::unique_ptr<rtc::Thread>>(threadNum);
    }

    rtc::Thread *GetThread() {
        // 先用读锁的方式从线程池中获取一个线程，如果此线程没有初始化，那么就用写锁的方式初始化它
        std::shared_lock<std::shared_mutex> sharedLock(threadsSharedMutex);
        if (!threads[threadIndex]) {
            sharedLock.unlock();
            std::unique_lock<std::shared_mutex> uniqueLock(threadsSharedMutex);
            if (!threads[threadIndex]) {
                threads[threadIndex] = rtc::Thread::Create();
                threads[threadIndex]->Start();
            }
            uniqueLock.unlock();
            sharedLock.lock();
        }

        // 指向下一个线程
        auto result = threads[threadIndex].get();
        threadIndex = (threadIndex + 1) % threads.size();

        // 返回线程
        return result;
    }

    rtc::Thread *GetSocketThread() {
        // 先用读锁的方式从线程池中获取一个线程，如果此线程没有初始化，那么就用写锁的方式初始化它
        std::shared_lock<std::shared_mutex> sharedLock(threadsSharedMutex);
        if (!socketThreads[threadIndex]) {
            sharedLock.unlock();
            std::unique_lock<std::shared_mutex> uniqueLock(threadsSharedMutex);
            if (!socketThreads[threadIndex]) {
                socketThreads[threadIndex] = rtc::Thread::CreateWithSocketServer();
                socketThreads[threadIndex]->Start();
            }
            uniqueLock.unlock();
            sharedLock.lock();
        }

        // 指向下一个线程
        auto result = socketThreads[threadIndex].get();
        threadIndex = (threadIndex + 1) % socketThreads.size();

        // 返回线程
        return result;
    }

    ~ThreadPool() {
        for (auto &thread : threads) {
            if (thread) {
                thread->Stop();
            }
        }
    }

  private:
    std::vector<std::unique_ptr<rtc::Thread>> threads;
    std::vector<std::unique_ptr<rtc::Thread>> socketThreads;
    std::shared_mutex threadsSharedMutex;
    std::atomic_uint32_t threadIndex = 0;
};

void *NewThreadPool(uint32_t threadNum) {
    return new ThreadPool(threadNum);
}

void *GetThreadPoolThread(void *threadPool) { return ((ThreadPool *)threadPool)->GetThread(); }

void *GetThreadPoolSocketThread(void *threadPool) {
    return ((ThreadPool *)threadPool)->GetSocketThread();
}

void DeleteThreadPool(void *threadPool) { delete (ThreadPool *)threadPool; }
