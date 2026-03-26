import { VM } from "vm2";

export interface PreScriptContext {
  headers: Record<string, string>;
  query: Record<string, string>;
  body: unknown;
}

export interface PostScriptContext {
  responseBody: unknown;
  responseHeaders: Record<string, string>;
}

export interface PreScriptResult {
  headers?: Record<string, string>;
  query?: Record<string, string>;
  body?: unknown;
}

export interface PostScriptResult {
  responseBody?: unknown;
  responseHeaders?: Record<string, string>;
}

export function executePreScript(
  script: string,
  context: PreScriptContext
): PreScriptResult {
  if (!script || script.trim() === "") {
    return {};
  }

  try {
    const vm = new VM({
      timeout: 5000,
      sandbox: {
        headers: JSON.parse(JSON.stringify(context.headers)),
        query: JSON.parse(JSON.stringify(context.query)),
        body: context.body ? JSON.parse(JSON.stringify(context.body)) : null,
        console: {
          log: (...args: unknown[]) => console.log("[PreScript]", ...args),
          error: (...args: unknown[]) => console.error("[PreScript]", ...args),
        },
        JSON,
        Object,
        Array,
        String,
        Number,
        Boolean,
        Math,
        Date,
        parseInt,
        parseFloat,
        isNaN,
        isFinite,
        encodeURIComponent,
        decodeURIComponent,
        decodeURI,
        encodeURI,
      },
    });

    const wrappedScript = `
      (function() {
        ${script}
        return {
          headers: typeof headers !== 'undefined' ? headers : undefined,
          query: typeof query !== 'undefined' ? query : undefined,
          body: typeof body !== 'undefined' ? body : undefined,
        };
      })()
    `;

    const result = vm.run(wrappedScript) as PreScriptResult;
    return result || {};
  } catch (error) {
    console.error("PreScript execution error:", error);
    throw new Error(`预处理脚本执行失败: ${error instanceof Error ? error.message : "未知错误"}`);
  }
}

export function executePostScript(
  script: string,
  context: PostScriptContext
): PostScriptResult {
  if (!script || script.trim() === "") {
    return {};
  }

  try {
    const vm = new VM({
      timeout: 5000,
      sandbox: {
        responseBody: context.responseBody
          ? JSON.parse(JSON.stringify(context.responseBody))
          : null,
        responseHeaders: JSON.parse(JSON.stringify(context.responseHeaders)),
        console: {
          log: (...args: unknown[]) => console.log("[PostScript]", ...args),
          error: (...args: unknown[]) => console.error("[PostScript]", ...args),
        },
        JSON,
        Object,
        Array,
        String,
        Number,
        Boolean,
        Math,
        Date,
        parseInt,
        parseFloat,
        isNaN,
        isFinite,
        encodeURIComponent,
        decodeURIComponent,
        decodeURI,
        encodeURI,
      },
    });

    const wrappedScript = `
      (function() {
        ${script}
        return {
          responseBody: typeof responseBody !== 'undefined' ? responseBody : undefined,
          responseHeaders: typeof responseHeaders !== 'undefined' ? responseHeaders : undefined,
        };
      })()
    `;

    const result = vm.run(wrappedScript) as PostScriptResult;
    return result || {};
  } catch (error) {
    console.error("PostScript execution error:", error);
    throw new Error(`后处理脚本执行失败: ${error instanceof Error ? error.message : "未知错误"}`);
  }
}
