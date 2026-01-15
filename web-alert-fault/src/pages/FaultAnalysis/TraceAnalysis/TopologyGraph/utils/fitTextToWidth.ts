// 获取字符串长度
export const getStringPxLength = (
  text: string,
  font = '12px Arial'
): number => {
  // 创建一个canvas元素用于测量文本
  const canvas = document.createElement('canvas');
  const context: any = canvas.getContext('2d');

  // 设置样式
  context.font = font;

  // 测量文本
  const metrics = context.measureText(text);

  return metrics.width;
};

export const fitTextToWidth = (
  text: string,
  maxWidth: number,
  font = '10px Arial'
): string => {
  if (getStringPxLength(text, font) <= maxWidth) {
    return text;
  }

  const ellipsis = '...';
  const ellipsisWidth = getStringPxLength(ellipsis, font);

  if (ellipsisWidth >= maxWidth) {
    return ellipsis;
  }

  const availableWidth = maxWidth - ellipsisWidth;
  let left = 0;
  let right = text.length;
  let result = text;

  while (left <= right) {
    const mid = Math.floor((left + right) / 2);
    const truncatedText = text.slice(0, mid);
    const truncatedWidth = getStringPxLength(truncatedText, font);

    if (truncatedWidth <= availableWidth) {
      result = truncatedText;
      left = mid + 1;
    } else {
      right = mid - 1;
    }
  }

  return `${result}${ellipsis}`;
};

export default fitTextToWidth;
