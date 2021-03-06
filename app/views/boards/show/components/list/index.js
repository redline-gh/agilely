import React, { useState, useRef } from "react";
import { authenticityToken, url, useOutsideAlerter } from "../lib";
import axios from "axios";
import VerticalList from "./vertical_list";
import HorizontalList from "./horizontal_list";

const List = (props) => {
  const [newCardIsOpen, toggleNewCard] = useState(false);
  const [newCardTitle, updateNewCardTitle] = useState("");
  const [title, setTitle] = useState(props.list.title);

  const newCardRef = useRef(null);
  useOutsideAlerter(newCardRef, () => toggleNewCard(false));

  const handleUpdateTitle = () => {
    event.preventDefault();
    if (title !== props.list.title) {
      axios.patch(`${url}/lists/${props.list._id.$oid}`, {
        authenticity_token: authenticityToken(),
        list: { title: title },
      });
    }
  };

  const handleNewCard = (title) => {
    event.preventDefault();
    if (/\S/.test(title)) {
      axios
        .post(`${url}/lists/${props.list._id.$oid}/cards`, {
          authenticity_token: authenticityToken(),
          card: { title: title },
        })
        .then((res) => {
          props.handleNewCard(props.list._id.$oid, res.data.card);
          toggleNewCard(false);
          updateNewCardTitle("");
        });
    }
  };
  if (props.view === "BOARD") {
    return (
      <VerticalList
        {...props}
        newCardIsOpen={newCardIsOpen}
        handleUpdateTitle={handleUpdateTitle}
        handleNewCard={handleNewCard}
        title={title}
        setTitle={setTitle}
        toggleNewCard={toggleNewCard}
        newCardRef={newCardRef}
        updateNewCardTitle={updateNewCardTitle}
        newCardTitle={newCardTitle}
      />
    );
  }
  else if (props.view === "LIST") {
    return (
      <HorizontalList
        {...props}
        newCardIsOpen={newCardIsOpen}
        handleUpdateTitle={handleUpdateTitle}
        handleNewCard={handleNewCard}
        title={title}
        setTitle={setTitle}
        toggleNewCard={toggleNewCard}
        newCardRef={newCardRef}
        updateNewCardTitle={updateNewCardTitle}
        newCardTitle={newCardTitle}
      />
    );
  }
};

export default List;
