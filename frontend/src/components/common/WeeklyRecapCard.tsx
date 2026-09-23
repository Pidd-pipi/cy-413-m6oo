import {Card,Col,Progress,Row,Statistic,Tag,Typography} from 'antd';import {useEffect,useState} from 'react';import {getWeeklyRecap} from '../../api/user';
import {MOOD_EMOJI,MOOD_LABELS} from '../../constants/mood';import {tagColor} from '../../utils/moodColor';
import type {MoodTag,WeeklyRecap} from '../../types';
export function WeeklyRecapCard(){const [recap,setRecap]=useState<WeeklyRecap>();useEffect(()=>{getWeeklyRecap().then(setRecap).catch(()=>{})},[]);if(!recap)return null;
const tag=recap.top_tag as MoodTag;
return <Card title="本周回顾" extra={<Typography.Text type="secondary">{recap.start_date} ~ {recap.end_date}</Typography.Text>}>
<Row gutter={16}>
<Col span={8}><Statistic title="情绪记录" value={recap.mood_count} suffix="条"/></Col>
<Col span={8}><Statistic title="平均心情" value={recap.average_mood??'—'} precision={recap.average_mood==null?undefined:1} suffix={recap.average_mood==null?'':'/10'}/></Col>
<Col span={8}><div className="ant-statistic-title">最常见标签</div><div className="ant-statistic-content">{tag?<Tag color={tagColor(tag)} style={{marginInlineEnd:0}}>{MOOD_EMOJI[tag]} {MOOD_LABELS[tag]}</Tag>:'—'}</div></Col>
<Col span={8} className="top-space"><Statistic title="日记" value={recap.journal_count} suffix="篇"/></Col>
<Col span={8} className="top-space"><Statistic title="测评完成" value={recap.assessment_done} suffix="次"/></Col>
<Col span={8} className="top-space"><Statistic title="距三次记录还差" value={recap.moods_to_three} suffix="次"/></Col>
</Row>
{recap.mood_count<3&&<Progress percent={Math.round(recap.mood_count/3*100)} size="small" className="top-space" showInfo={false}/>}
{recap.mood_count===0&&<p className="muted top-space">本周还没有记录，从今天的第一条心情开始吧。</p>}
</Card>}
