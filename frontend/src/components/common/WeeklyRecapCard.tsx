import {Card,Col,Progress,Row,Statistic,Tag} from 'antd';
import type {MoodTag,WeeklyRecap} from '../../types';
import {MOOD_EMOJI,MOOD_LABELS} from '../../constants/mood';

const GOAL=3;
export function WeeklyRecapCard({recap}:{recap?:WeeklyRecap}){
  const percent=recap?Math.min(100,Math.round(recap.mood_count/GOAL*100)):0;
  const topTag=(recap?.top_tag||'') as MoodTag|'';
  const range=recap?`${recap.start_date} 至 ${recap.end_date}`:'—';
  return <Card title={`本周回顾（${range}）`} className="top-space">
    <Row gutter={[16,16]}>
      <Col xs={12} sm={8}><Statistic title="情绪记录" value={recap?.mood_count??0} suffix="条"/></Col>
      <Col xs={12} sm={8}><Statistic title="平均心情" value={recap?.average_mood??'—'} precision={recap?.average_mood==null?undefined:1} suffix={recap?.average_mood==null?'':' / 10'}/></Col>
      <Col xs={12} sm={8}>
        <div className="ant-statistic-title">最常见标签</div>
        {topTag?<Tag color="blue">{MOOD_EMOJI[topTag]} {MOOD_LABELS[topTag]}</Tag>:<span className="muted">—</span>}
      </Col>
      <Col xs={12} sm={8}><Statistic title="日记篇数" value={recap?.journal_count??0} suffix="篇"/></Col>
      <Col xs={12} sm={8}><Statistic title="测评完成" value={recap?.assessment_count??0} suffix="次"/></Col>
      <Col xs={12} sm={8}><Statistic title="距三次记录还差" value={recap?.moods_to_three??GOAL} suffix="次"/></Col>
    </Row>
    <Progress percent={percent} showInfo={false} className="top-space"/>
    <p className="muted">{recap&&recap.moods_to_three===0?'本周三次记录目标已达成，继续保持！':'本周至少记录 3 次心情，帮你看清情绪变化。'}</p>
  </Card>;
}
