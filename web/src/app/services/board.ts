import { BoardApi } from '../../features/board/adapter/api/boardApi';
import { BoardUseCases } from '../../features/board/usecase/boardUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    board: BoardUseCases;
  }
}

export const service = defineService('board', (http) => new BoardUseCases(new BoardApi(http)));
