package com.anjing.model.response;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.Collections;
import java.util.List;

/**
 * 标准分页 payload。
 *
 * <p>字段必须与 platform contract 保持一致：records/current/size/total。</p>
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class PageResult<T> {

    private List<T> records;

    private Integer current;

    private Integer size;

    private Long total;

    public static <T> PageResult<T> of(List<T> records, long total, int current, int size) {
        return new PageResult<>(
                records == null ? Collections.emptyList() : records,
                current,
                size,
                total
        );
    }

    public static <T> PageResult<T> empty(int current, int size) {
        return of(Collections.emptyList(), 0L, current, size);
    }
}
