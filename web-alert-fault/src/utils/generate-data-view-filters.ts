import {
  MetricModelQueryParams,
  PaginationParams
} from '@/constants/commonTypes';

/**
 * 数据视图过滤器以及排序参数转换
 * @param params
 * @returns
 */
export const generateDataViewFilters = (params: any) => {
  const {
    sorter = {},
    filters = {},
    likeFields = [],
    extraOriginFilters = []
  } = params;
  const formattedParams: Pick<PaginationParams, 'filters' | 'sort'> = {};
  const filterKeys = Object.keys(filters).filter((key) =>
    Array.isArray(filters[key]) ? filters[key].length : filters[key]
  );

  if (filterKeys.length === 0) {
    if (extraOriginFilters.length > 0) {
      formattedParams.filters =
        extraOriginFilters.length === 1
          ? extraOriginFilters[0]
          : {
              operation: 'and',
              sub_conditions: extraOriginFilters
            };
    }
  } else if (filterKeys.length === 1) {
    const isMultiValue = Array.isArray(filters[filterKeys[0]]);

    if (extraOriginFilters.length > 0) {
      formattedParams.filters = {
        operation: 'and',
        sub_conditions: [
          ...extraOriginFilters,
          {
            field: filterKeys[0],
            operation: likeFields.includes(filterKeys[0])
              ? ('like' as const)
              : isMultiValue
              ? 'in'
              : 'match_phrase',
            value: filters[filterKeys[0]],
            value_from: 'const'
          }
        ]
      };
    } else {
      formattedParams.filters = {
        field: filterKeys[0],
        operation: likeFields.includes(filterKeys[0])
          ? ('like' as const)
          : isMultiValue
          ? 'in'
          : 'match_phrase',
        value: filters[filterKeys[0]],
        value_from: 'const'
      };
    }
  } else if (filterKeys.length > 1) {
    const sub_conditions = extraOriginFilters
      .concat(
        filterKeys.map((key) => ({
          field: key,
          operation: likeFields.includes(key)
            ? ('like' as const)
            : Array.isArray(filters[key])
            ? ('in' as const)
            : ('match_phrase' as const),
          value: filters[key],
          value_from: 'const' as const
        }))
      )
      .filter(Boolean);

    formattedParams.filters = {
      operation: 'and',
      sub_conditions
    };
  }

  if (sorter.field && sorter.order) {
    formattedParams.sort = [
      {
        field: sorter.field,
        direction: sorter.order === 'ascend' ? 'asc' : 'desc'
      }
    ];
  }

  return formattedParams;
};

export const generateMetricModelFilters = (filters: any) => {
  const formattedParams: Pick<MetricModelQueryParams, 'filters'> = {};

  const filterKeys = Object.keys(filters).filter((key) => filters[key].length);

  const conditions = filterKeys.map((key) => ({
    name: key,
    operation: Array.isArray(filters[key]) ? ('in' as const) : ('=' as const),
    value: filters[key]
  }));

  formattedParams.filters = conditions;

  return formattedParams;
};
