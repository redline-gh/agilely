class Participation < ApplicationRecord
  belongs_to :participant, polymorphic: true
  belongs_to :user

  validates :participant_id, uniqueness: { scope: [:participant_type, :user_id] }
end