class User < ApplicationRecord
  has_many :participations
  has_many :boards, through: :participations, source: :participant, source_type: "Board", dependent: :destroy

  devise :database_authenticatable, :registerable,
         :recoverable, :rememberable, :validatable,
         :confirmable

  validates :password_confirmation, presence: true, on: :create
  validates :name, presence: true, length: { maximum: 50 }
end