class BoardsController < ApplicationController

  def create
    board = Boards::CreateService.execute(board_params: board_params, user: current_user)
    redirect_to show_board_path(slug: board.slug, data: { turbolinks: false })
  end

  def show
    @board = authorize Board.full(params[:slug])
    @board_titles = current_user&.board_titles
    @role = current_user&.role_in @board || "guest"
  end

  def update
    Board.find_by(slug: params[:slug]).update(board_params)
    head :no_content
  end

  def destroy
    Board.find_by(slug: params[:slug]).destroy
    head :no_content
  end

  private

  def board_params
    params.require(:board).permit(:title, :public)
  end
end
