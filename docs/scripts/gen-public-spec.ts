import { readFileSync, writeFileSync } from 'node:fs'
import { resolve, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const ROOT = resolve(__dirname, '../../')

// 外部開発者向けに公開するパスと許可するメソッドの定義
const PUBLIC_PATHS: Record<string, string[]> = {
  '/idols': ['get'],
  '/idols/{id}': ['get'],
  '/idols/{id}/external-ids': ['get'],
  '/groups': ['get'],
  '/groups/{id}': ['get'],
  '/agencies': ['get'],
  '/agencies/{id}': ['get'],
  '/events': ['get'],
  '/events/upcoming': ['get'],
  '/events/{id}': ['get'],
  '/releases': ['get'],
  '/releases/{id}': ['get'],
  '/tags': ['get'],
  '/tags/{id}': ['get'],
  '/billing/checkout-sessions': ['post'],
  '/billing/portal-sessions': ['post'],
}

type OpenAPISpec = {
  info: Record<string, unknown>
  host: string
  basePath: string
  schemes: string[]
  paths: Record<string, Record<string, unknown>>
  definitions: Record<string, unknown>
  securityDefinitions?: Record<string, unknown>
  [key: string]: unknown
}

const srcPath = resolve(ROOT, 'backend/docs/swagger.json')
const destPath = resolve(__dirname, '../public/openapi.json')

const full: OpenAPISpec = JSON.parse(readFileSync(srcPath, 'utf-8'))

// 公開パス・メソッドだけを抽出
const filteredPaths: Record<string, Record<string, unknown>> = {}
for (const [path, allowedMethods] of Object.entries(PUBLIC_PATHS)) {
  if (!full.paths[path]) continue
  const ops: Record<string, unknown> = {}
  for (const method of allowedMethods) {
    if (full.paths[path][method]) {
      ops[method] = full.paths[path][method]
    }
  }
  if (Object.keys(ops).length > 0) {
    filteredPaths[path] = ops
  }
}

// 使用されている $ref を収集して definitions を絞り込む
const specStr = JSON.stringify(filteredPaths)
const usedDefs = new Set<string>()
const refRegex = /"#\/definitions\/([^"]+)"/g
let m: RegExpExecArray | null
while ((m = refRegex.exec(specStr)) !== null) {
  usedDefs.add(m[1])
}

const filteredDefs: Record<string, unknown> = {}
for (const defName of usedDefs) {
  if (full.definitions[defName]) {
    filteredDefs[defName] = full.definitions[defName]
  }
}

const publicSpec: OpenAPISpec = {
  swagger: '2.0',
  info: {
    ...full.info,
    title: 'Idol API',
    description:
      'アイドル・グループ・事務所・イベント情報を提供する公開 REST API。' +
      'API キーは /billing/checkout-sessions からご取得ください。',
  },
  host: process.env.API_HOST ?? 'localhost:8081',
  basePath: full.basePath,
  schemes: process.env.API_HOST ? ['https'] : ['http'],
  consumes: ['application/json'],
  produces: ['application/json'],
  securityDefinitions: full.securityDefinitions,
  paths: filteredPaths,
  definitions: filteredDefs,
}

writeFileSync(destPath, JSON.stringify(publicSpec, null, 2), 'utf-8')
console.log(`Generated: ${destPath} (${Object.keys(filteredPaths).length} paths)`)
