use ratatui::{
    style::{Color, Modifier, Style},
    text::{Line, Span},
    widgets::{List, ListItem},
};
use regex::Regex;

pub struct App {
    pub room: String,
    pub messages: Vec<String>,
}

impl App {
    pub fn new() -> Self {
        App {
            room: "zero".to_owned(),
            messages: Vec::new(),
        }
    }

    pub fn messages_to_list(&self, min_lines: usize, _max_len: usize) -> List<'_> {
        let msgex = Regex::new(r#"^\w+:\s"#).unwrap();
        let mut list_items = Vec::new();
        let a = self.messages.to_owned();
        for msg in a.into_iter().rev() {
            let display_line = if msg.starts_with("***") {
                Line::from(vec![
                    Span::styled(
                        msg,
                        Style::default()
                            .fg(Color::Blue)
                            .add_modifier(Modifier::ITALIC | Modifier::BOLD),
                    ),
                    Span::from("\n"), // padding
                ])
                .centered()
            } else if msgex.is_match(&msg) {
                let parts = msg
                    .splitn(2, ' ')
                    .map(|p| p.to_string())
                    .collect::<Vec<String>>();
                Line::from(vec![
                    Span::styled(
                        format!("{} ", parts[0]),
                        Style::default()
                            .fg(Color::Green)
                            .add_modifier(Modifier::BOLD),
                    ),
                    Span::from(parts[1].clone()),
                ])
            } else if msg.starts_with("!!! ") {
                let parts = msg
                    .splitn(2, '.')
                    .map(|p| p.to_string())
                    .collect::<Vec<String>>();
                Line::from(vec![
                    Span::styled(
                        format!("{}.", parts[0]),
                        Style::default().fg(Color::Red).add_modifier(Modifier::BOLD),
                    ),
                    Span::styled(
                        parts[1].clone(),
                        Style::default()
                            .fg(Color::LightGreen)
                            .add_modifier(Modifier::ITALIC),
                    ),
                ])
                .centered()
            } else {
                Line::styled(
                    msg,
                    Style::default()
                        .fg(Color::Cyan)
                        .add_modifier(Modifier::ITALIC),
                )
                .centered()
            };
            list_items.push(ListItem::new(display_line));
            if list_items.len() >= min_lines {
                break;
            }
        }
        while list_items.len() < min_lines {
            list_items.push(ListItem::new(""));
        }
        list_items.reverse();
        List::new(list_items)
    }

    pub fn add_msg(&mut self, msg: String) {
        self.messages.push(msg);
    }

    pub fn update_room_name(&mut self, new_room: &str) {
        new_room.clone_into(&mut self.room);
    }
}
