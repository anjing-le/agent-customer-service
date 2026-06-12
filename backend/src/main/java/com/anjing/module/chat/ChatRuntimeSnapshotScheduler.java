package com.anjing.module.chat;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

/**
 * Chat Runtime 快照定时采样。
 */
@Slf4j
@Component
@RequiredArgsConstructor
@ConditionalOnProperty(prefix = "app.customer-service.runtime-snapshot", name = "enabled", havingValue = "true")
public class ChatRuntimeSnapshotScheduler {

    private final ChatService chatService;

    @Scheduled(cron = "${app.customer-service.runtime-snapshot.cron:0 5 0 * * *}")
    public void captureDailySnapshot() {
        ChatVO.RuntimeSnapshotVO snapshot = chatService.captureRuntimeSnapshot("scheduled");
        log.info("Chat Runtime 快照采样完成: snapshotId={}, date={}", snapshot.getSnapshotId(), snapshot.getSnapshotDate());
    }
}
