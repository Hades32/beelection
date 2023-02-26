export const rateLimit = <
  T extends (...args: any) => void,
  P extends Parameters<T>,
>(
  func: T,
  period: number
): ((...args: P) => void) => {
  let timer: NodeJS.Timeout;
  let pendingArgs: P;
  return function (...args: P) {
    const context = this;
    if (!timer) {
      func.apply(context, args);
      timer = setTimeout(() => {
        timer = null;
        if (pendingArgs) {
          func.apply(context, pendingArgs);
          pendingArgs = null;
        }
      }, period);
    } else {
      pendingArgs = args;
    }
  };
};
