/// ビットマスクを扱うユーティリティ関数を定義するファイル

/**
 * @summary ビットマスクを結合する関数
 * @description 複数のビットマスクをOR演算で結合して1つのビットマスクにする関数
 * @param bitmasks 結合するビットマスクの配列
 * @returns 結合されたビットマスク
 * @author yuito-it <yuito@uniproject.jp>
 **/
export const MergeBitmask = (bitmasks: bigint[]): bigint => {
  // ビットマスクをOR演算で結合して1つのビットマスクにする
  return bitmasks.reduce((acc, bitmask) => acc | bitmask, 0n);
};

export const IsIncludedInBitmask = (
  bitmask: bigint,
  target: bigint,
): boolean => {
  // targetのビットがbitmaskに含まれているかを判定する
  if (target === 0n) {
    // targetが0の場合は常にfalseを返す
    return false;
  }
  if (target === -1n) {
    return true;
  }
  return (bitmask & target) === target;
};
