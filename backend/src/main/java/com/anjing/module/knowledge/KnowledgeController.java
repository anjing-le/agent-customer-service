package com.anjing.module.knowledge;

import com.anjing.model.constants.ApiConstants;
import com.anjing.model.response.APIResponse;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

/**
 * 知识中心 Controller
 * 负责处理各类知识（商品、活动、FAQ、行业、解决方案）的管理
 */
@RestController
@RequiredArgsConstructor
public class KnowledgeController {

    private final KnowledgeService knowledgeService;

    // ==================== 知识总览 ====================

    /**
     * 获取知识总览
     */
    @PostMapping(ApiConstants.Knowledge.OVERVIEW)
    public APIResponse<KnowledgeVO.KnowledgeOverviewVO> getOverview() {
        return APIResponse.success(knowledgeService.getOverview());
    }

    // ==================== 商品知识 ====================

    /**
     * 获取商品知识列表
     */
    @PostMapping(ApiConstants.Knowledge.PRODUCT_LIST)
    public APIResponse<KnowledgeVO.PageVO<KnowledgeVO.ProductVO>> listProducts(@RequestBody KnowledgeDTO.QueryProductDTO dto) {
        return APIResponse.success(knowledgeService.listProducts(dto));
    }

    /**
     * 创建商品知识
     */
    @PostMapping(ApiConstants.Knowledge.PRODUCT_CREATE)
    public APIResponse<KnowledgeVO.ProductVO> createProduct(@RequestBody KnowledgeDTO.CreateProductDTO dto) {
        return APIResponse.success(knowledgeService.createProduct(dto));
    }

    /**
     * 更新商品知识
     */
    @PostMapping(ApiConstants.Knowledge.PRODUCT_UPDATE)
    public APIResponse<KnowledgeVO.ProductVO> updateProduct(@RequestBody KnowledgeDTO.UpdateProductDTO dto) {
        return APIResponse.success(knowledgeService.updateProduct(dto));
    }

    /**
     * 删除商品知识
     */
    @PostMapping(ApiConstants.Knowledge.PRODUCT_DELETE)
    public APIResponse<Void> deleteProduct(@RequestBody KnowledgeDTO.IdDTO dto) {
        knowledgeService.deleteProduct(dto.getId());
        return APIResponse.success();
    }

    /**
     * 获取商品知识详情
     */
    @PostMapping(ApiConstants.Knowledge.PRODUCT_DETAIL)
    public APIResponse<KnowledgeVO.ProductVO> getProductDetail(@RequestBody KnowledgeDTO.IdDTO dto) {
        return APIResponse.success(knowledgeService.getProductDetail(dto.getId()));
    }

    // ==================== 活动知识 ====================

    /**
     * 获取活动知识列表
     */
    @PostMapping(ApiConstants.Knowledge.ACTIVITY_LIST)
    public APIResponse<KnowledgeVO.PageVO<KnowledgeVO.ActivityVO>> listActivities(@RequestBody KnowledgeDTO.QueryActivityDTO dto) {
        return APIResponse.success(knowledgeService.listActivities(dto));
    }

    /**
     * 创建活动知识
     */
    @PostMapping(ApiConstants.Knowledge.ACTIVITY_CREATE)
    public APIResponse<KnowledgeVO.ActivityVO> createActivity(@RequestBody KnowledgeDTO.CreateActivityDTO dto) {
        return APIResponse.success(knowledgeService.createActivity(dto));
    }

    /**
     * 更新活动知识
     */
    @PostMapping(ApiConstants.Knowledge.ACTIVITY_UPDATE)
    public APIResponse<KnowledgeVO.ActivityVO> updateActivity(@RequestBody KnowledgeDTO.UpdateActivityDTO dto) {
        return APIResponse.success(knowledgeService.updateActivity(dto));
    }

    /**
     * 删除活动知识
     */
    @PostMapping(ApiConstants.Knowledge.ACTIVITY_DELETE)
    public APIResponse<Void> deleteActivity(@RequestBody KnowledgeDTO.IdDTO dto) {
        knowledgeService.deleteActivity(dto.getId());
        return APIResponse.success();
    }

    /**
     * 获取活动知识详情
     */
    @PostMapping(ApiConstants.Knowledge.ACTIVITY_DETAIL)
    public APIResponse<KnowledgeVO.ActivityVO> getActivityDetail(@RequestBody KnowledgeDTO.IdDTO dto) {
        return APIResponse.success(knowledgeService.getActivityDetail(dto.getId()));
    }

    // ==================== FAQ问答 ====================

    /**
     * 获取FAQ列表
     */
    @PostMapping(ApiConstants.Knowledge.FAQ_LIST)
    public APIResponse<KnowledgeVO.PageVO<KnowledgeVO.FaqVO>> listFaqs(@RequestBody KnowledgeDTO.QueryFaqDTO dto) {
        return APIResponse.success(knowledgeService.listFaqs(dto));
    }

    /**
     * 创建FAQ
     */
    @PostMapping(ApiConstants.Knowledge.FAQ_CREATE)
    public APIResponse<KnowledgeVO.FaqVO> createFaq(@RequestBody KnowledgeDTO.CreateFaqDTO dto) {
        return APIResponse.success(knowledgeService.createFaq(dto));
    }

    /**
     * 更新FAQ
     */
    @PostMapping(ApiConstants.Knowledge.FAQ_UPDATE)
    public APIResponse<KnowledgeVO.FaqVO> updateFaq(@RequestBody KnowledgeDTO.UpdateFaqDTO dto) {
        return APIResponse.success(knowledgeService.updateFaq(dto));
    }

    /**
     * 删除FAQ
     */
    @PostMapping(ApiConstants.Knowledge.FAQ_DELETE)
    public APIResponse<Void> deleteFaq(@RequestBody KnowledgeDTO.IdDTO dto) {
        knowledgeService.deleteFaq(dto.getId());
        return APIResponse.success();
    }

    /**
     * 获取FAQ详情
     */
    @PostMapping(ApiConstants.Knowledge.FAQ_DETAIL)
    public APIResponse<KnowledgeVO.FaqVO> getFaqDetail(@RequestBody KnowledgeDTO.IdDTO dto) {
        return APIResponse.success(knowledgeService.getFaqDetail(dto.getId()));
    }

    // ==================== 知识缺口 ====================

    /**
     * 获取知识缺口运营统计
     */
    @PostMapping(ApiConstants.Knowledge.GAP_SUMMARY)
    public APIResponse<KnowledgeVO.KnowledgeGapSummaryVO> getKnowledgeGapSummary() {
        return APIResponse.success(knowledgeService.getKnowledgeGapSummary());
    }

    /**
     * 获取运行时知识缺口列表
     */
    @PostMapping(ApiConstants.Knowledge.GAP_LIST)
    public APIResponse<KnowledgeVO.PageVO<KnowledgeVO.KnowledgeGapVO>> listKnowledgeGaps(@RequestBody KnowledgeDTO.QueryGapDTO dto) {
        return APIResponse.success(knowledgeService.listKnowledgeGaps(dto));
    }

    /**
     * 处理知识缺口，可补成 FAQ 或仅标记已处理
     */
    @PostMapping(ApiConstants.Knowledge.GAP_RESOLVE)
    public APIResponse<KnowledgeVO.KnowledgeGapVO> resolveKnowledgeGap(@RequestBody KnowledgeDTO.ResolveGapDTO dto) {
        return APIResponse.success(knowledgeService.resolveKnowledgeGap(dto));
    }

    // ==================== 行业知识 ====================

    /**
     * 获取行业知识列表
     */
    @PostMapping(ApiConstants.Knowledge.INDUSTRY_LIST)
    public APIResponse<KnowledgeVO.PageVO<KnowledgeVO.IndustryVO>> listIndustries(@RequestBody KnowledgeDTO.QueryIndustryDTO dto) {
        return APIResponse.success(knowledgeService.listIndustries(dto));
    }

    /**
     * 创建行业知识
     */
    @PostMapping(ApiConstants.Knowledge.INDUSTRY_CREATE)
    public APIResponse<KnowledgeVO.IndustryVO> createIndustry(@RequestBody KnowledgeDTO.CreateIndustryDTO dto) {
        return APIResponse.success(knowledgeService.createIndustry(dto));
    }

    /**
     * 更新行业知识
     */
    @PostMapping(ApiConstants.Knowledge.INDUSTRY_UPDATE)
    public APIResponse<KnowledgeVO.IndustryVO> updateIndustry(@RequestBody KnowledgeDTO.UpdateIndustryDTO dto) {
        return APIResponse.success(knowledgeService.updateIndustry(dto));
    }

    /**
     * 删除行业知识
     */
    @PostMapping(ApiConstants.Knowledge.INDUSTRY_DELETE)
    public APIResponse<Void> deleteIndustry(@RequestBody KnowledgeDTO.IdDTO dto) {
        knowledgeService.deleteIndustry(dto.getId());
        return APIResponse.success();
    }

    // ==================== 场景解决方案 ====================

    /**
     * 获取解决方案列表
     */
    @PostMapping(ApiConstants.Knowledge.SOLUTION_LIST)
    public APIResponse<KnowledgeVO.PageVO<KnowledgeVO.SolutionVO>> listSolutions(@RequestBody KnowledgeDTO.QuerySolutionDTO dto) {
        return APIResponse.success(knowledgeService.listSolutions(dto));
    }

    /**
     * 创建解决方案
     */
    @PostMapping(ApiConstants.Knowledge.SOLUTION_CREATE)
    public APIResponse<KnowledgeVO.SolutionVO> createSolution(@RequestBody KnowledgeDTO.CreateSolutionDTO dto) {
        return APIResponse.success(knowledgeService.createSolution(dto));
    }

    /**
     * 更新解决方案
     */
    @PostMapping(ApiConstants.Knowledge.SOLUTION_UPDATE)
    public APIResponse<KnowledgeVO.SolutionVO> updateSolution(@RequestBody KnowledgeDTO.UpdateSolutionDTO dto) {
        return APIResponse.success(knowledgeService.updateSolution(dto));
    }

    /**
     * 删除解决方案
     */
    @PostMapping(ApiConstants.Knowledge.SOLUTION_DELETE)
    public APIResponse<Void> deleteSolution(@RequestBody KnowledgeDTO.IdDTO dto) {
        knowledgeService.deleteSolution(dto.getId());
        return APIResponse.success();
    }

    /**
     * 获取解决方案详情
     */
    @PostMapping(ApiConstants.Knowledge.SOLUTION_DETAIL)
    public APIResponse<KnowledgeVO.SolutionVO> getSolutionDetail(@RequestBody KnowledgeDTO.IdDTO dto) {
        return APIResponse.success(knowledgeService.getSolutionDetail(dto.getId()));
    }

    // ==================== 知识向量化 ====================

    /**
     * 触发知识向量化
     */
    @PostMapping(ApiConstants.Knowledge.VECTORIZE)
    public APIResponse<KnowledgeVO.VectorizeResultVO> vectorize(@RequestBody KnowledgeDTO.VectorizeDTO dto) {
        return APIResponse.success(knowledgeService.vectorize(dto));
    }

    /**
     * 查询向量化状态
     */
    @PostMapping(ApiConstants.Knowledge.VECTORIZE_STATUS)
    public APIResponse<KnowledgeVO.VectorizeStatusVO> getVectorizeStatus(@RequestBody KnowledgeDTO.VectorizeStatusDTO dto) {
        return APIResponse.success(knowledgeService.getVectorizeStatus(dto));
    }

    // ==================== 知识导入导出 ====================

    /**
     * 导入商品知识
     */
    @PostMapping(ApiConstants.Knowledge.PRODUCT_IMPORT)
    public APIResponse<KnowledgeVO.ImportResultVO> importProducts(@RequestBody KnowledgeDTO.ImportDTO dto) {
        return APIResponse.success(knowledgeService.importKnowledge(dto));
    }

    /**
     * 导出商品知识
     */
    @PostMapping(ApiConstants.Knowledge.PRODUCT_EXPORT)
    public APIResponse<KnowledgeVO.ExportResultVO> exportProducts(@RequestBody KnowledgeDTO.ExportDTO dto) {
        return APIResponse.success(knowledgeService.exportKnowledge(dto));
    }

    /**
     * 导入FAQ
     */
    @PostMapping(ApiConstants.Knowledge.FAQ_IMPORT)
    public APIResponse<KnowledgeVO.ImportResultVO> importFaqs(@RequestBody KnowledgeDTO.ImportDTO dto) {
        return APIResponse.success(knowledgeService.importKnowledge(dto));
    }
}
