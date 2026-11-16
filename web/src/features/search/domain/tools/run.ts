import type { ToolEnv, ToolName } from './detect';
import type { ColourResult } from './colour';
import type { PasswordResult } from './password';
import type { SubnetResult } from './subnet';
import type { TimestampResult } from './timestamp';
import type { UuidResult } from './uuid';

export type ToolResult = PasswordResult | UuidResult | ColourResult | TimestampResult | SubnetResult;

/**
 * Runs a tool, importing its module on first use (#477). Each import is its
 * own chunk, so a tool costs nothing until someone types a query for it.
 */
export async function runTool(name: ToolName, query: string, env: ToolEnv): Promise<ToolResult | null> {
  switch (name) {
    case 'password':
      return (await import('./password')).password(query, env);
    case 'uuid':
      return (await import('./uuid')).uuid(query, env);
    case 'colour':
      return (await import('./colour')).colour(query);
    case 'timestamp':
      return (await import('./timestamp')).timestamp(query, env);
    case 'subnet':
      return (await import('./subnet')).subnet(query);
  }
}
