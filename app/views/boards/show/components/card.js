import React, { useState } from "react";
import { Draggable } from "react-beautiful-dnd";
import CardDetail from "./card_detail";

const Card = (props) => {
  const [modalIsOpen, openModal] = useState(false);

  return (
    <div className="cursor-pointer">
      <Draggable
        draggableId={props.card._id.$oid}
        index={props.index}
        isDragDisabled={!props.can_edit}
      >
        {(provided) => (
          <div
            ref={provided.innerRef}
            {...provided.draggableProps}
            {...provided.dragHandleProps}
            className="mt-2"
            onClick={() => openModal(true)}
          >
            <div className="py-2 px-3 bg-white rounded-md shadow flex flex-wrap justify-between items-baseline">
              <span
                className="text-sm font-normal leading-snug text-gray-900 break-all"
                id={`card-${props.card.id}-title`}
              >
                {props.card.title}
              </span>
              <span className="text-xs text-gray-600">
                {props.card.position}
              </span>
            </div>
          </div>
        )}
      </Draggable>
      {modalIsOpen && <CardDetail />}
    </div>
  );
};

export default Card;
