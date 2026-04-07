export const toCamelcase = <T>(obj: any): T => {
  if (Array.isArray(obj)) {
    return obj.map((item) => toCamelcase(item)) as unknown as T;
  } else if (obj !== null && typeof obj === "object") {
    const newObj: any = {};
    for (const key of Object.keys(obj)) {
      const camelKey = key.replace(/_([a-z])/g, (_, letter) =>
        letter.toUpperCase(),
      );
      newObj[camelKey] = toCamelcase(obj[key]);
    }
    return newObj as T;
  }
  return obj as T;
};

export const toSnakecase = <T>(obj: any): T => {
  if (Array.isArray(obj)) {
    return obj.map((item) => toSnakecase(item)) as unknown as T;
  } else if (obj !== null && typeof obj === "object") {
    const newObj: any = {};
    for (const key of Object.keys(obj)) {
      const snakeKey = key.replace(/([A-Z])/g, "_$1").toLowerCase();
      newObj[snakeKey] = toSnakecase(obj[key]);
    }
    return newObj as T;
  }
  return obj as T;
};
