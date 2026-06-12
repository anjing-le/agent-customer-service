package com.anjing.module.agent.runtime;

import com.anjing.module.agent.application.KnowledgeRetriever;
import com.anjing.module.agent.domain.ConversationTurn;
import com.anjing.module.agent.domain.IntentAnalysis;
import com.anjing.module.agent.domain.KnowledgeEvidence;
import com.anjing.module.agent.domain.KnowledgeRecall;
import com.anjing.module.agent.domain.KnowledgeSource;
import com.anjing.module.knowledge.entity.Activity;
import com.anjing.module.knowledge.entity.Faq;
import com.anjing.module.knowledge.entity.Product;
import com.anjing.module.knowledge.repository.ActivityRepository;
import com.anjing.module.knowledge.repository.FaqRepository;
import com.anjing.module.knowledge.repository.ProductRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.Collections;
import java.util.Comparator;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 当前 V2 runtime 的 JPA 知识检索实现。
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class JpaKnowledgeRetriever implements KnowledgeRetriever {

    private static final double MATCH_THRESHOLD = 0.3;
    private static final int MAX_EVIDENCE_PER_KIND = 3;

    private final ProductRepository productRepository;
    private final ActivityRepository activityRepository;
    private final FaqRepository faqRepository;

    @Override
    public KnowledgeRecall retrieve(ConversationTurn turn, IntentAnalysis analysis) {
        KnowledgeRecall recall = new KnowledgeRecall();
        recall.setMinAcceptedScore(MATCH_THRESHOLD);

        List<Long> selectedProductIds = parseIds(turn.getContext(), "selectedProducts");
        List<Long> selectedActivityIds = parseIds(turn.getContext(), "selectedActivities");

        recall.getEvidences().addAll(retrieveSelectedProducts(selectedProductIds));
        recall.getEvidences().addAll(retrieveSelectedActivities(selectedActivityIds));

        if (selectedProductIds.isEmpty()) {
            recall.getEvidences().addAll(retrieveProductsByKeyword(turn.getUserMessage()));
        }
        if (selectedActivityIds.isEmpty() && containsAny(turn.getUserMessage(), "优惠", "活动", "打折", "促销", "价格", "便宜")) {
            recall.getEvidences().addAll(retrieveActivitiesByTrigger());
        }
        recall.getEvidences().addAll(retrieveFaqsByKeyword(turn.getUserMessage()));

        log.info("Agent 知识检索完成: evidenceCount={}, reliable={}",
                recall.getEvidences().size(), recall.hasReliableEvidence());
        return recall;
    }

    private List<KnowledgeEvidence> retrieveSelectedProducts(List<Long> ids) {
        if (ids.isEmpty()) return Collections.emptyList();
        List<KnowledgeEvidence> evidences = new ArrayList<>();
        for (Long id : ids) {
            productRepository.findById(id).ifPresent(product ->
                    evidences.add(productEvidence(product, 0.99, "人工选择")));
        }
        return evidences;
    }

    private List<KnowledgeEvidence> retrieveSelectedActivities(List<Long> ids) {
        if (ids.isEmpty()) return Collections.emptyList();
        List<KnowledgeEvidence> evidences = new ArrayList<>();
        for (Long id : ids) {
            activityRepository.findById(id).ifPresent(activity ->
                    evidences.add(activityEvidence(activity, 0.99, "人工选择")));
        }
        return evidences;
    }

    private List<KnowledgeEvidence> retrieveProductsByKeyword(String query) {
        List<KnowledgeEvidence> evidences = new ArrayList<>();
        for (Product product : productRepository.findAll()) {
            double score = calculateMatchScore(query, product.getProductName(), product.getDescription(), product.getFeatures());
            if (score > MATCH_THRESHOLD) {
                evidences.add(productEvidence(product, score, "关键词匹配"));
            }
        }
        return topByScore(evidences);
    }

    private List<KnowledgeEvidence> retrieveActivitiesByTrigger() {
        List<KnowledgeEvidence> evidences = new ArrayList<>();
        for (Activity activity : activityRepository.findAll()) {
            evidences.add(activityEvidence(activity, 0.85, "关键词触发"));
        }
        return topByScore(evidences);
    }

    private List<KnowledgeEvidence> retrieveFaqsByKeyword(String query) {
        List<KnowledgeEvidence> evidences = new ArrayList<>();
        for (Faq faq : faqRepository.findAll()) {
            double score = calculateMatchScore(query, faq.getQuestion(), faq.getAnswer(), faq.getRelatedQuestions());
            if (score > MATCH_THRESHOLD) {
                evidences.add(faqEvidence(faq, score));
            }
        }
        return topByScore(evidences);
    }

    private KnowledgeEvidence productEvidence(Product product, double score, String matchMode) {
        KnowledgeEvidence evidence = new KnowledgeEvidence();
        evidence.setEvidenceId(String.valueOf(product.getId()));
        evidence.setSource(KnowledgeSource.PRODUCT);
        evidence.setTitle(product.getProductName());
        evidence.setContent(product.getDescription());
        evidence.setScore(score);
        evidence.setAttributes(attributes(
                "productName", product.getProductName(),
                "matchMode", matchMode
        ));
        return evidence;
    }

    private KnowledgeEvidence activityEvidence(Activity activity, double score, String matchMode) {
        KnowledgeEvidence evidence = new KnowledgeEvidence();
        evidence.setEvidenceId(String.valueOf(activity.getId()));
        evidence.setSource(KnowledgeSource.ACTIVITY);
        evidence.setTitle(activity.getActivityName());
        evidence.setContent(activity.getDescription());
        evidence.setScore(score);
        evidence.setAttributes(attributes(
                "activityName", activity.getActivityName(),
                "description", activity.getDescription(),
                "matchMode", matchMode
        ));
        return evidence;
    }

    private KnowledgeEvidence faqEvidence(Faq faq, double score) {
        KnowledgeEvidence evidence = new KnowledgeEvidence();
        evidence.setEvidenceId(String.valueOf(faq.getId()));
        evidence.setSource(KnowledgeSource.FAQ);
        evidence.setTitle(faq.getQuestion());
        evidence.setContent(faq.getAnswer());
        evidence.setScore(score);
        evidence.setAttributes(attributes(
                "question", faq.getQuestion(),
                "answer", faq.getAnswer()
        ));
        return evidence;
    }

    private Map<String, Object> attributes(Object... pairs) {
        Map<String, Object> attributes = new HashMap<>();
        for (int i = 0; i + 1 < pairs.length; i += 2) {
            attributes.put(String.valueOf(pairs[i]), pairs[i + 1]);
        }
        return attributes;
    }

    private List<KnowledgeEvidence> topByScore(List<KnowledgeEvidence> evidences) {
        evidences.sort(Comparator.comparing(KnowledgeEvidence::getScore, Comparator.nullsLast(Double::compareTo)).reversed());
        if (evidences.size() <= MAX_EVIDENCE_PER_KIND) return evidences;
        return new ArrayList<>(evidences.subList(0, MAX_EVIDENCE_PER_KIND));
    }

    private double calculateMatchScore(String query, String... fields) {
        if (query == null || query.isEmpty()) return 0;
        String q = query.toLowerCase();
        int totalHits = 0;
        int totalChars = 0;
        for (String field : fields) {
            if (field == null || field.isEmpty()) continue;
            String f = field.toLowerCase();
            totalChars += f.length();
            for (int i = 0; i < q.length(); i++) {
                if (f.indexOf(q.charAt(i)) >= 0) totalHits++;
            }
            for (int len = 2; len <= Math.min(q.length(), 6); len++) {
                for (int i = 0; i <= q.length() - len; i++) {
                    if (f.contains(q.substring(i, i + len))) totalHits += len;
                }
            }
        }
        if (totalChars == 0) return 0;
        return Math.min(1.0, totalHits / (double) (q.length() * 3 + totalChars * 0.5));
    }

    private List<Long> parseIds(Map<String, Object> context, String key) {
        if (context == null || !context.containsKey(key)) return Collections.emptyList();
        Object value = context.get(key);
        if (!(value instanceof List<?> values)) return Collections.emptyList();

        List<Long> result = new ArrayList<>();
        for (Object item : values) {
            if (item instanceof Number number) {
                result.add(number.longValue());
            } else if (item instanceof String text) {
                try {
                    result.add(Long.parseLong(text));
                } catch (NumberFormatException ignored) {
                    log.debug("忽略无法解析的知识 ID: {}", text);
                }
            }
        }
        return result;
    }

    private boolean containsAny(String content, String... keywords) {
        if (content == null) return false;
        for (String keyword : keywords) {
            if (content.contains(keyword)) return true;
        }
        return false;
    }
}
