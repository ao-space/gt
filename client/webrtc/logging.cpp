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

#include <mutex>

#include <rtc_base/logging.h>

#include "logging.h"

using namespace std;
using namespace rtc;

class LogSink : public rtc::LogSink {
  protected:
    void OnLogMessage(const string &message, LoggingSeverity severity, const char *tag) {
        auto messageStr = (string)message;
        if (messageStr.back() = '\n') {
            messageStr.pop_back();
        }
        if (messageStr.back() = '\r') {
            messageStr.pop_back();
        }
        onLogMessage(severity, (char *)messageStr.c_str(), (char *)tag);
    }

    void OnLogMessage(const string &message, LoggingSeverity severity) {
        auto messageStr = (string)message;
        if (messageStr.back() = '\n') {
            messageStr.pop_back();
        }
        if (messageStr.back() = '\r') {
            messageStr.pop_back();
        }
        onLogMessage(severity, (char *)messageStr.c_str(), nullptr);
    }

    void OnLogMessage(const string &message) {
        auto messageStr = (string)message;
        if (messageStr.back() = '\n') {
            messageStr.pop_back();
        }
        if (messageStr.back() = '\r') {
            messageStr.pop_back();
        }
        onLogMessage(rtc::LS_INFO, (char *)messageStr.c_str(), nullptr);
    }

    void OnLogMessage(absl::string_view message, LoggingSeverity severity, const char *tag) {
        auto messageStr = (string)message;
        if (messageStr.back() = '\n') {
            messageStr.pop_back();
        }
        if (messageStr.back() = '\r') {
            messageStr.pop_back();
        }
        onLogMessage(severity, (char *)messageStr.c_str(), (char *)tag);
    }

    void OnLogMessage(absl::string_view message, LoggingSeverity severity) {
        auto messageStr = (string)message;
        if (messageStr.back() = '\n') {
            messageStr.pop_back();
        }
        if (messageStr.back() = '\r') {
            messageStr.pop_back();
        }
        onLogMessage(severity, (char *)messageStr.c_str(), nullptr);
    }

    void OnLogMessage(absl::string_view message) {
        auto messageStr = (string)message;
        if (messageStr.back() = '\n') {
            messageStr.pop_back();
        }
        if (messageStr.back() = '\r') {
            messageStr.pop_back();
        }
        onLogMessage(rtc::LS_INFO, (char *)messageStr.c_str(), nullptr);
    }
};

void SetLog(int severity) {
    static mutex m;
    m.lock();
    static ::LogSink *stream = nullptr;
    if (stream != nullptr) {
        LogMessage::RemoveLogToStream(stream);
        delete stream;
    }
    stream = new ::LogSink();
    LogMessage::AddLogToStream(stream, LoggingSeverity(severity));
    m.unlock();
}
